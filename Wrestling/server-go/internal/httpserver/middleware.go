package httpserver

import (
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"wrestling/internal/handlers"
	"wrestling/internal/observability"
)

func Stack(logger *log.Logger, metrics *observability.Metrics, origins string, next http.Handler) http.Handler {
	return handlers.EnableCORSWithOrigins(origins,
		SecurityRateLimit(
			Recovery(logger,
				BodyLimit(2<<20,
					handlers.RequestID(
						handlers.SecurityHeaders(
							handlers.RequestLogger(logger, measure(metrics, next)),
						),
					),
				),
			),
		),
	)
}

func measure(metrics *observability.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)
		metrics.Observe(sw.status, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Write(p []byte) (int, error) { return w.ResponseWriter.Write(p) }

func Recovery(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Printf("panic request_id=%s error=%v\n%s", w.Header().Get("X-Request-ID"), v, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func BodyLimit(max int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, max)
		}
		next.ServeHTTP(w, r)
	})
}

type bucket struct {
	count int
	reset time.Time
}
type limiter struct {
	mu     sync.Mutex
	items  map[string]bucket
	limit  int
	window time.Duration
}

func newLimiter(limit int, window time.Duration) *limiter {
	return &limiter{items: map[string]bucket{}, limit: limit, window: window}
}
func (l *limiter) allow(key string) (bool, time.Duration) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.items[key]
	if !ok || now.After(b.reset) {
		b = bucket{reset: now.Add(l.window)}
	}
	b.count++
	l.items[key] = b
	if b.count > l.limit {
		return false, time.Until(b.reset)
	}
	if len(l.items) > 10000 {
		for k, v := range l.items {
			if now.After(v.reset) {
				delete(l.items, k)
			}
		}
	}
	return true, 0
}

func SecurityRateLimit(next http.Handler) http.Handler {
	general := newLimiter(300, time.Minute)
	auth := newLimiter(12, time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		lim := general
		if strings.HasPrefix(r.URL.Path, "/api/auth/") {
			lim = auth
		}
		ok, wait := lim.allow(key)
		if !ok {
			seconds := int(wait.Seconds()) + 1
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// Do not trust X-Forwarded-For by default; configure a trusted reverse proxy before using it.
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
