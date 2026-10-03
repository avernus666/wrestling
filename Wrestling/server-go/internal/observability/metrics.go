package observability

import (
	"expvar"
	"net/http"
	"sync/atomic"
	"time"
)

type Metrics struct {
	requests   atomic.Uint64
	errors     atomic.Uint64
	totalNanos atomic.Uint64
}

func NewMetrics() *Metrics { return &Metrics{} }
func (m *Metrics) Observe(status int, elapsed time.Duration) {
	m.requests.Add(1)
	m.totalNanos.Add(uint64(elapsed.Nanoseconds()))
	if status >= 500 {
		m.errors.Add(1)
	}
}
func (m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	expvar.Do(func(kv expvar.KeyValue) { _, _ = w.Write([]byte(kv.Key + " " + kv.Value.String() + "\n")) })
	_, _ = w.Write([]byte("wrestling_requests_total " + itoa(m.requests.Load()) + "\n"))
	_, _ = w.Write([]byte("wrestling_errors_total " + itoa(m.errors.Load()) + "\n"))
	_, _ = w.Write([]byte("wrestling_request_nanos_total " + itoa(m.totalNanos.Load()) + "\n"))
}
func itoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	b := [20]byte{}
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
