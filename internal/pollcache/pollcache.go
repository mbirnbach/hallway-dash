// Package pollcache implements a background-refreshed cache: a fetch
// function runs once immediately and then on a fixed interval, and the
// last successful result is served instantly to HTTP handlers. A failed
// fetch just logs and keeps serving the previous value, so a flaky
// upstream API degrades to "stale but present" rather than an error
// bubbling up to the wall display.
package pollcache

import (
	"context"
	"log"
	"sync"
	"time"
)

type Cache[T any] struct {
	mu       sync.RWMutex
	val      T
	name     string
	interval time.Duration
	fetch    func(ctx context.Context) (T, error)
}

func New[T any](name string, interval time.Duration, fetch func(ctx context.Context) (T, error)) *Cache[T] {
	return &Cache[T]{name: name, interval: interval, fetch: fetch}
}

// Start fetches immediately and then refreshes on the configured
// interval until ctx is cancelled.
func (c *Cache[T]) Start(ctx context.Context) {
	c.refresh(ctx)
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.refresh(ctx)
			}
		}
	}()
}

func (c *Cache[T]) refresh(ctx context.Context) {
	fetchCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	v, err := c.fetch(fetchCtx)
	if err != nil {
		log.Printf("[%s] refresh failed: %v", c.name, err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *Cache[T]) Get() T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}
