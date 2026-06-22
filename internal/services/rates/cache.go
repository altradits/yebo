package rates

import (
	"sync"
	"time"
)

// Cache holds the latest BTC/KES rate in memory.
type Cache struct {
	mu        sync.RWMutex
	btcKES    float64
	btcUSD    float64
	fetchedAt time.Time
}

var Global = &Cache{}

func (c *Cache) Set(btcKES, btcUSD float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.btcKES = btcKES
	c.btcUSD = btcUSD
	c.fetchedAt = time.Now()
}

func (c *Cache) GetKES() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.btcKES
}

func (c *Cache) GetUSD() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.btcUSD
}

// GetBTCKES is a convenience wrapper for use in handlers.
func GetBTCKES() (float64, bool) {
	v := Global.GetKES()
	return v, v > 0
}

func (c *Cache) Age() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.fetchedAt.IsZero() {
		return 0
	}
	return time.Since(c.fetchedAt)
}
