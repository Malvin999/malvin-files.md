package sync

import (
	"fmt"
	"sync"
	"time"
)

// PerUserLocker is a structure that is used to ensure that only one request per requestInfo is processed at a time
type PerUserLocker interface {
	// Lock locks the requestInfo's mutex and remembers the request
	// If the requestInfo's mutex is already locked, Lock blocks until the mutex is unlocked
	// userID is a unique requestInfo identifier
	// request is anything representation of the request, it's used for logging and debugging
	Lock(userID int64, request interface{})
	// TryLock is the same as Lock, but it does not block if the requestInfo's mutex is already locked
	// It returns true if the mutex was locked and false otherwise
	TryLock(userID int64, request interface{}) bool
	// Unlock unlocks(releases) the requestInfo's mutex
	// if the mutex is not locked, Unlock panics
	Unlock(userID int64)
	// Len returns the number requests that are currently being processed
	Len() int
	// FrozenRequests returns a map of requests that are being processed longer than `threshold` and were started after `since`
	FrozenRequests(threshold time.Duration, since time.Time) map[int64]interface{}
}

type requestInfo struct {
	mutex       *sync.Mutex
	request     interface{} // the request that is being currently processed, used for logging and debugging
	requestsInQ int         // number of requests that are waiting for the mutex to be unlocked
	startedAt   time.Time   // time when the request was started
}

func NewPerUserLocker() PerUserLocker {
	return &locker{
		requests: make(map[int64]*requestInfo),
	}
}

type locker struct {
	requests map[int64]*requestInfo
	mapLock  sync.Mutex
}

func (l *locker) TryLock(userID int64, request interface{}) bool {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	u := l.getUser(userID)
	if u.mutex.TryLock() {
		u.requestsInQ++
		u.request = request
		u.startedAt = time.Now()
		return true
	}
	return false
}

func (l *locker) Lock(userID int64, request interface{}) {
	l.mapLock.Lock()
	u := l.getUser(userID)
	u.requestsInQ++
	l.mapLock.Unlock()
	u.mutex.Lock()
	u.request = request
	u.startedAt = time.Now()
}

func (l *locker) Unlock(userID int64) {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	user, ok := l.requests[userID]
	if !ok {
		panic(fmt.Sprintf("Unlocking non-existing mutex, userID=%v", userID))
	}
	user.mutex.Unlock()
	user.requestsInQ--
	if user.requestsInQ == 0 {
		delete(l.requests, userID)
	}
}

func (l *locker) Len() int {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	return len(l.requests)
}

func (l *locker) FrozenRequests(threshold time.Duration, since time.Time) map[int64]interface{} {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	res := make(map[int64]interface{})
	for userID, u := range l.requests {
		if time.Since(u.startedAt) < threshold {
			continue // the request is not frozen yet
		}
		if u.startedAt.Before(since) {
			continue // the request is too old
		}
		res[userID] = u.request
	}
	return res
}

// getUser returns requestInfo by userID, if requestInfo does not exist, it creates new one
// the function is not thread safe, it should be with locked mapLock
func (l *locker) getUser(userID int64) *requestInfo {
	u, ok := l.requests[userID]
	if !ok {
		u = &requestInfo{
			mutex: &sync.Mutex{},
		}
		l.requests[userID] = u
	}
	return u
}
