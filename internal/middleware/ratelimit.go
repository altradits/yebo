package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// bucket is a simple token bucket.
type bucket struct {
	tokens   float64
	last     time.Time
	capacity float64
	rate     float64 // tokens per second
	mu       sync.Mutex
}

func newBucket(capacity float64, perSecond float64) *bucket {
	return &bucket{tokens: capacity, last: time.Now(), capacity: capacity, rate: perSecond}
}

func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// IPRateLimit limits requests per IP at the given rate (requests/minute).
func IPRateLimit(perMinute float64) func(http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		buckets = make(map[string]*bucket)
	)
	go func() {
		for range time.Tick(5 * time.Minute) {
			mu.Lock()
			buckets = make(map[string]*bucket)
			mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			mu.Lock()
			b, ok := buckets[ip]
			if !ok {
				b = newBucket(perMinute, perMinute/60.0)
				buckets[ip] = b
			}
			mu.Unlock()
			if !b.allow() {
				http.Error(w, "Rate limit exceeded. Please wait.", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
