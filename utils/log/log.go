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

	maxBody = 1 << 14 // 16KiB
)

// TraceHandler allows the slog to add a trace ID to logs from the context.
// To add a trace ID to the context, use TraceCtxKey:
//
//	ctx = context.WithValue(ctx, log.TraceCtxKey, 'put_trace_id_here')
type TraceHandler struct {
	slog.Handler
}

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(TraceCtxKey).(string); ok {
		r.Add("trace_id", slog.StringValue(traceID))
	}

	return h.Handler.Handle(ctx, r)
}

// DumpHttpRequest dumps the HTTP request and prints out.
func DumpHttpRequest(ctx context.Context, r *http.Request, level slog.Level) {
	dumpFunc := httputil.DumpRequestOut
	if r.URL.Scheme == "" || r.URL.Host == "" {
		dumpFunc = httputil.DumpRequest
	}
	b, err := dumpFunc(r, r.ContentLength < maxBody)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP REQUEST", "error", err)
		return
	}
	slog.Log(ctx, level, "HTTP REQUEST", "dump", string(b))
}

// DumpHttpResponse dumps the HTTP response and prints out.
func DumpHttpResponse(ctx context.Context, r *http.Response, level slog.Level) {
	b, err := httputil.DumpResponse(r, r.ContentLength < maxBody)
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
