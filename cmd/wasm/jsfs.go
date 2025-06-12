package main

import (
	"fmt"
	"os"
	"syscall/js"
	"time"

	"github.com/spf13/afero"
)

type JSFS struct{}

func NewJSFS() *JSFS {
	return &JSFS{}
}

func (fs *JSFS) Create(name string) (afero.File, error) {
	return nil, nil
}

func (fs *JSFS) Mkdir(name string, perm os.FileMode) error {
	resultChan := make(chan struct{}, 1)
	errorChan := make(chan error, 1)

	callAsync("mkdir", func(result js.Value, err error) {
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- struct{}{}
	}, name)

	select {
	case result := <-resultChan:
		sendToJS(fmt.Sprintf("%v", result))
		return nil
	case err := <-errorChan:
		return err
	}
}

func (fs *JSFS) MkdirAll(path string, perm os.FileMode) error {
	resultChan := make(chan struct{}, 1)
	errorChan := make(chan error, 1)

	callAsync("mkdirAll", func(result js.Value, err error) {
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- struct{}{}
	}, path)

	select {
	case result := <-resultChan:
		sendToJS(fmt.Sprintf("%v", result))
		return nil
	case err := <-errorChan:
		return err
	}
}

func (fs *JSFS) Open(name string) (afero.File, error) {
	return nil, nil
}

func (fs *JSFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, nil
}

func (fs *JSFS) Remove(name string) error {
	return nil
}

func (fs *JSFS) RemoveAll(path string) error {
	return nil
}

func (fs *JSFS) Rename(oldname, newname string) error {
	return nil
}

func (fs *JSFS) Stat(name string) (os.FileInfo, error) {
	return nil, nil
}

func (fs *JSFS) Name() string {
	return ""
}

func (fs *JSFS) Chmod(name string, mode os.FileMode) error {
	return nil
}

func (fs *JSFS) Chown(name string, uid, gid int) error {
	return nil
}

func (fs *JSFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}

func readFile(_ afero.Fs, path string) ([]byte, error) {
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	callAsync("read", func(result js.Value, err error) {
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result.String()
	}, path)

	select {
	case result := <-resultChan:
		sendToJS(result)
		return []byte(result), nil
	case err := <-errorChan:
		return nil, err
	}
}

func writeFile(_ afero.Fs, path string, data []byte, perm os.FileMode) error {
	resultChan := make(chan struct{}, 1)
	errorChan := make(chan error, 1)

	callAsync("write", func(result js.Value, err error) {
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- struct{}{}
	}, path, string(data))

	select {
	case result := <-resultChan:
		sendToJS(result)
		return nil
	case err := <-errorChan:
		return err
	}
}

func exists(_ afero.Fs, path string) (bool, error) {
	resultChan := make(chan bool, 1)
	errorChan := make(chan error, 1)

	callAsync("exists", func(result js.Value, err error) {
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result.Bool()
	}, path)

	select {
	case result := <-resultChan:
		sendToJS(result)
		return result, nil
	case err := <-errorChan:
		return false, err
	}
}
