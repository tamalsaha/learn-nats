package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/tamalsaha/learn-nats/shared"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
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
//
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
	buf.Reset()

	if err := r.WriteProxy(buf); err != nil {
		return nil, err
	}

	timeout := rt.timeout(r.Context(), time.Now())

	r2 := R{
		Request:            buf.Bytes(),
		TLS:                rt.TLS,
		Timeout:            max(0, timeout-500*time.Millisecond),
		DisableCompression: rt.DisableCompression,
	}
	buf.Reset()
	if err := json.NewEncoder(buf).Encode(r2); err != nil {
		return nil, err
	}

	return Proxy(r, rt.Conn, rt.Subject, buf.Bytes(), timeout)
}

const HeaderKeyDone = "Done"

// SEE: https://github.com/nats-io/nats.docs/blob/master/using-nats/developing-with-nats/sending/replyto.md#including-a-reply-subject
func Proxy(req *http.Request, nc *nats.Conn, reSubj string, data []byte, timeout time.Duration) (*http.Response, error) {
	hubRespSub, edgeRespSub := shared.ProxyResponseSubjects()

	// Listen for a single response
	sub, err := nc.SubscribeSync(hubRespSub)
	if err != nil {
		return nil, err
	}

	// Send the request.
	if err := nc.PublishRequest(reSubj, edgeRespSub, data); err != nil {
		return nil, err
	}

	r, w := io.Pipe()
	go func() {
		var e2 error

		defer func() {
			if e2 != nil {
				_ = w.CloseWithError(e2)
			} else {
				_ = w.Close()
			}
			_ = sub.Unsubscribe()
		}()

		for {
			var msg *nats.Msg
			msg, e2 = sub.NextMsg(timeout)
			if e2 != nil {
				if e2 == nats.ErrTimeout {
					e2 = nil
					continue // ignore ErrTimeout
				}
				break
			}

			_, e2 = w.Write(msg.Data)
			if e2 != nil {
				break
			}
			if results, ok := msg.Header[HeaderKeyDone]; ok {
				if results[0] != "" {
					e2 = errors.New(results[0])
				}
				break
			}
		}
	}()

	return http.ReadResponse(bufio.NewReader(r), req)
}
