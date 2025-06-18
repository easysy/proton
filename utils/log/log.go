package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
)

type contextKey int

const (
	TraceCtxKey contextKey = iota + 1

	traceLogKey = "trace_id"
)

var (
	traceKey = traceLogKey
)

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

// DumpHttpRequest logs the full HTTP request using slog at the specified log level.
// It uses DumpRequestOut for client requests or DumpRequest for server requests,
// and includes the body if its size is less than maxBody.
func DumpHttpRequest(ctx context.Context, r *http.Request, level slog.Level, maxBody int64) {
	dumpFunc := httputil.DumpRequestOut
	if r.URL.Scheme == "" || r.URL.Host == "" {
		dumpFunc = httputil.DumpRequest
	}

	b, err := dumpFunc(r, r.ContentLength < maxBody)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP REQUEST", "error", err)
		return
	}

	if r.Body != nil && maxBody > 0 {
		var lBody []byte
		if r.Body, lBody, err = bodyReader(r.Body, maxBody); err != nil {
			slog.ErrorContext(ctx, "HTTP REQUEST", "error", err)
			return
		}
		b = append(b, lBody...)
	}

	slog.Log(ctx, level, "HTTP REQUEST", "dump", string(b))
}

// DumpHttpResponse logs the full HTTP response using slog at the specified log level.
// It includes the body if its size is less than maxBody.
func DumpHttpResponse(ctx context.Context, r *http.Response, level slog.Level, maxBody int64) {
	b, err := httputil.DumpResponse(r, r.ContentLength < maxBody)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP RESPONSE", "error", err)
		return
	}

	if r.Body != nil && maxBody > 0 {
		var limitedCopy []byte
		if r.Body, limitedCopy, err = bodyReader(r.Body, maxBody); err != nil {
			slog.ErrorContext(ctx, "HTTP RESPONSE", "error", err)
			return
		}
		b = append(b, limitedCopy...)
	}

	slog.Log(ctx, level, "HTTP RESPONSE", "dump", string(b))
}

func bodyReader(body io.ReadCloser, limit int64) (io.ReadCloser, []byte, error) {
	defer Closer(nil, body)

	fullCopy, limitedCopy := new(bytes.Buffer), new(bytes.Buffer)
	limitReader := io.TeeReader(io.LimitReader(body, limit), limitedCopy)

	n, err := io.Copy(fullCopy, io.MultiReader(limitReader, body))
	if err != nil {
		return nil, nil, err
	}

	if n > limit {
		limitedCopy.WriteString("...\n--- the body is limited due to size limit ---")
	}

	return io.NopCloser(fullCopy), limitedCopy.Bytes(), nil
}

// Closer calls the Close method, if the closure occurred with an error, it prints out.
func Closer(ctx context.Context, c io.Closer) {
	if err := c.Close(); err != nil {
		slog.ErrorContext(ctx, "close resource", "error", err)
	}
}
