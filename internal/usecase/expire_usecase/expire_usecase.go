package expire_usecase

import (
	"sync"
	"time"
)

type ExpirerInterface interface {
	SetExpiration(Key string, duration time.Duration, callback func())
	IsExpired(key string) bool
	ExpireKey(key string)
}

type DefaultExpirer struct {
	expiredKeys map[string]bool
	mu          sync.Mutex
}

func NewDefaultExpirer() *DefaultExpirer {
	return &DefaultExpirer{
		expiredKeys: make(map[string]bool),
	}
}

func (e *DefaultExpirer) SetExpiration(key string, duration time.Duration, callback func()) {
	time.AfterFunc(duration, func() {
		e.ExpireKey(key)
		if callback != nil {
			callback()
		}
	})
}

func (e *DefaultExpirer) IsExpired(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.expiredKeys[key]
}

func (e *DefaultExpirer) ExpireKey(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.expiredKeys[key] = true
}
