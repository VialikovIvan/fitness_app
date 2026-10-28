package middleware

import (
	"hash/fnv"
	"net"
	"net/http"
	"sync"
)

func (m *Manager) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			m.logger.Warn("failed to split IP and port", "IP", r.RemoteAddr)
			userIP = r.RemoteAddr
		}

		mu := m.getLockByIP(userIP)

		mu.Lock()
		allowed, err := m.limiter.AllowRequestByIP(r.Context(), userIP)
		mu.Unlock()

		if err != nil {
			m.logger.Error("rate limiting error", "err", err)
		}
		if !allowed {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Manager) getLockByIP(ip string) *sync.Mutex {
	h := fnv.New64a()
	h.Write([]byte(ip))
	return &m.locks[h.Sum64()%numLocks]
}
