// Package contact provides the site-bound email capability for the Contact app.
package contact

import (
	"sync"
	"time"
)

type bucket struct {
	count int
	until time.Time
}

// Budget bounds contact mail across all sites, independently of login mail.
// There is no queue or disk spool.
type Budget struct {
	mu     sync.Mutex
	window bucket
	active int
}

func (b *Budget) acquire(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !now.Before(b.window.until) {
		b.window = bucket{until: now.Add(time.Hour)}
	}
	if b.active >= 4 || b.window.count >= 200 {
		return false
	}
	b.active++
	b.window.count++
	return true
}

func (b *Budget) release() { b.mu.Lock(); b.active--; b.mu.Unlock() }
