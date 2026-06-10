package hutil

import (
	"sync"
	"time"
)

type TokenBlacklist struct {
	mu    sync.RWMutex
	items map[string]time.Time
}

var TokenBl *TokenBlacklist

func InitTokenBlacklist() {
	TokenBl = &TokenBlacklist{
		items: make(map[string]time.Time),
	}
	go TokenBl.cleanupLoop(5 * time.Minute)
}

func (b *TokenBlacklist) Add(jti string, exp time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[jti] = exp
}

func (b *TokenBlacklist) IsRevoked(jti string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, exists := b.items[jti]
	return exists
}

func (b *TokenBlacklist) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.items)
}

func (b *TokenBlacklist) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		b.mu.Lock()
		for jti, exp := range b.items {
			if now.After(exp) {
				delete(b.items, jti)
			}
		}
		b.mu.Unlock()
	}
}
