package httpapi

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"github.com/example/go-redis-rate-limiter/internal/limiter"
)

func RateLimit(rl *limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := r.Header.Get("X-Client-ID")
			if clientID == "" {
				clientID = remoteIP(r.RemoteAddr)
			}

			algorithm := strings.ToLower(r.Header.Get("X-Rate-Limit-Algorithm"))
			if algorithm != "sliding" {
				algorithm = "fixed"
			}

			result, err := rl.Allow(r.Context(), clientID, algorithm)
			if err != nil {
				http.Error(w, `{"error":"rate limiter unavailable"}`, http.StatusServiceUnavailable)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(result.Reset.Seconds())))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(max(1, int(result.Reset.Seconds()))))
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func remoteIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

	func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
