package feishu

import (
	"hash/fnv"
	"math"
)

type identityMapper struct {
	defaultUserID int64
}

func newIdentityMapper(defaultUserID int64) identityMapper {
	return identityMapper{defaultUserID: defaultUserID}
}

func (m identityMapper) UserID(openID string) int64 {
	if m.defaultUserID > 0 {
		return m.defaultUserID
	}

	h := fnv.New64a()
	_, _ = h.Write([]byte(openID))
	return int64(h.Sum64() & math.MaxInt64)
}

func allowlist(values []string) map[string]bool {
	if len(values) == 0 {
		return nil
	}

	allowed := make(map[string]bool, len(values))
	for _, value := range values {
		if value != "" {
			allowed[value] = true
		}
	}
	return allowed
}
