package log

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
)

type contextKey int

const (
	TraceCtxKey contextKey = iota + 1

	maxBody     int64 = 1 << 14 // 16KiB
	traceLogKey       = "trace_id"
)

var (
	body     = maxBody
	traceKey = traceLogKey
)

// SetMaxBodyLen sets the maximum body length for HTTP request/response dumping.
// The default limit is 16KiB.
// If a positive length is provided, and it is less than the default, it updates the limit to the specified value.
func SetMaxBodyLen(length int64) {
	body = maxBody
	if length > 0 && length < maxBody {
		body = length
	}
}

// SetTraceKey sets the key used to store trace IDs in log records.
// If a non-empty key is provided, it overrides the default trace key ("trace_id").
func SetTraceKey(key string) {
	traceKey = traceLogKey
	if key != "" {
		traceKey = key
	}
}

// TraceHandler allows the slog to add a trace ID to logs from the context.
// To add a trace ID to the context, use TraceCtxKey:
//
//	ctx = context.WithValue(ctx, log.TraceCtxKey, 'your_id_here')
type TraceHandler struct {
	slog.Handler
}

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(TraceCtxKey).(string); ok {
		r.Add(traceKey, slog.StringValue(traceID))
	}

	return h.Handler.Handle(ctx, r)
}

// DumpHttpRequest dumps the HTTP request and prints out.
func DumpHttpRequest(ctx context.Context, r *http.Request, level slog.Level) {
	dumpFunc := httputil.DumpRequestOut
	if r.URL.Scheme == "" || r.URL.Host == "" {
		dumpFunc = httputil.DumpRequest
	}
	b, err := dumpFunc(r, r.ContentLength < body)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP REQUEST", "error", err)
		return
	}
	slog.Log(ctx, level, "HTTP REQUEST", "dump", string(b))
}

// DumpHttpResponse dumps the HTTP response and prints out.
func DumpHttpResponse(ctx context.Context, r *http.Response, level slog.Level) {
	b, err := httputil.DumpResponse(r, r.ContentLength < body)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP RESPONSE", "error", err)
		return
	}
	slog.Log(ctx, level, "HTTP RESPONSE", "dump", string(b))
}

// Closer calls the Close method, if the closure occurred with an error, it prints out.
func Closer(ctx context.Context, c io.Closer) {
	if err := c.Close(); err != nil {
		slog.ErrorContext(ctx, "close resource", "error", err)
	}
}
