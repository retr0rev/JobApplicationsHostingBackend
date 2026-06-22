package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	b        int
	ttl      time.Duration
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	irl := &IPRateLimiter{
		limiters: make(map[string]*ipLimiter),
		r:        r,
		b:        b,
		ttl:      10 * time.Minute,
	}
	go irl.gcLoop()
	return irl
}

func (i *IPRateLimiter) gcLoop() {
	t := time.NewTicker(i.ttl)
	defer t.Stop()
	for range t.C {
		i.mu.Lock()
		cutoff := time.Now().Add(-i.ttl)
		for ip, l := range i.limiters {
			if l.lastSeen.Before(cutoff) {
				delete(i.limiters, ip)
			}
		}
		i.mu.Unlock()
	}
}

func (i *IPRateLimiter) get(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()
	l, ok := i.limiters[ip]
	if !ok {
		l = &ipLimiter{limiter: rate.NewLimiter(i.r, i.b)}
		i.limiters[ip] = l
	}
	l.lastSeen = time.Now()
	return l.limiter
}

func (i *IPRateLimiter) Allow(ip string) bool {
	return i.get(ip).Allow()
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func RateLimit(l *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !l.Allow(ip) {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				log.Printf("rate limit: ip=%s method=%s path=%s", ip, r.Method, r.URL.Path)
				http.Error(w, `{"error":"too many requests, please try again later"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
