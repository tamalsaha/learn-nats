package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
	"time"
)

type NatsTransport struct {
	Conn    *nats.Conn
	Subject string
	Timeout time.Duration
	// DisableCompression bypasses automatic GZip compression requests to the
	// server.
	DisableCompression bool
	TLS                *PersistableTLSConfig
}

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func max(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}

func min(x, y time.Duration) time.Duration {
	if x < y {
		return x
	}
	return y
}

const defaultTimeout = 30 * time.Second // from http.DefaultTransport

// timeout returns the minimum of:
//   - Timeout
//   - the context's deadline-now
// Or defaultTimeout, if none of Timeout, or context's deadline-now is set.
func (rt *NatsTransport) timeout(ctx context.Context, now time.Time) time.Duration {
	timeout := rt.Timeout
	if d, ok := ctx.Deadline(); ok {
		timeout = min(timeout, d.Sub(now))
	}
	if timeout > 0 {
		return timeout
	}
	return defaultTimeout
}

func (rt *NatsTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	buf := pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)

	if err := r.Write(buf); err != nil { // WriteProxy
		return nil, err
	}

	timeout := rt.timeout(r.Context(), time.Now())

	r2 := R{
		Request: buf.Bytes(),
		TLS:     rt.TLS,
		Timeout: max(0, timeout-500*time.Millisecond),
		DisableCompression: rt.DisableCompression, // transport.Config
	}
	buf.Reset()
	if err := json.NewEncoder(buf).Encode(r2); err != nil {
		return nil, err
	}

	msg, err := rt.Conn.RequestMsg(&nats.Msg{
		Subject: rt.Subject,
		Data:    buf.Bytes(),
	}, timeout)
	if err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(bytes.NewReader(msg.Data)), r)
}
