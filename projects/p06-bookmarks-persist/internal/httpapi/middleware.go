package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// RequestIDHeader je hlavička identifikátoru požadavku.
const RequestIDHeader = "X-Request-ID"

type ctxKey int

const requestIDKey ctxKey = 1

var errPanic = errors.New("panic v handleru")

// Options konfiguruje middleware kolem routeru.
type Options struct {
	MaxBodyBytes int64
	Timeout      time.Duration
}

// Wrap složí middleware: recovery → request ID → body limit → timeout → handler.
func Wrap(h http.Handler, opts Options) http.Handler {
	next := h
	if opts.Timeout > 0 {
		next = http.TimeoutHandler(next, opts.Timeout, "timeout")
	}
	if opts.MaxBodyBytes > 0 {
		next = bodyLimit(next, opts.MaxBodyBytes)
	}
	next = requestID(next)
	return recovery(next)
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeError(w, errPanic)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func bodyLimit(next http.Handler, max int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > max {
			w.Header().Set("Content-Type", ProblemContentType)
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_, _ = w.Write([]byte(`{"type":"about:blank","title":"Příliš velké tělo","status":413}` + "\n"))
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, max)
		}
		next.ServeHTTP(w, r)
	})
}
