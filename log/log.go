package log

import (
	"context"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httputil"
	"strings"
)

type contextKey int

const (
	TraceCtxKey contextKey = iota + 1
	traceLogKey            = "trace_id"
	holder                 = "<binary body>\r\n"
)

var (
	traceKey = traceLogKey
)

// SetTraceKey sets the key used to store trace IDs in log records.
// If a non-empty key is provided, it overrides the default trace key ("trace_id").
// It must be called before any concurrent use of DumpHttpRequest or DumpHttpResponse.
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

var types = map[string]struct{}{
	"json":                  {},
	"ld+json":               {},
	"problem+json":          {},
	"xml":                   {},
	"atom+xml":              {},
	"problem+xml":           {},
	"rss+xml":               {},
	"soap+xml":              {},
	"xhtml+xml":             {},
	"graphql":               {},
	"javascript":            {},
	"x-javascript":          {},
	"x-www-form-urlencoded": {},
}

// SetHumanReadableSubTypes overrides the list of application/* subtypes whose bodies are logged as text.
// It must be called before any concurrent use of DumpHttpRequest or DumpHttpResponse.
func SetHumanReadableSubTypes(list []string) {
	if len(list) == 0 {
		return
	}
	types = make(map[string]struct{}, len(list))
	for _, t := range list {
		types[t] = struct{}{}
	}
}

// DumpHttpRequest logs the full HTTP request using slog at the specified log level.
// It uses DumpRequestOut for client requests and DumpRequest for server requests.
// If body is true and the Content-Type is human-readable, the body is included in the dump.
// If the body is binary, a placeholder is appended instead.
func DumpHttpRequest(ctx context.Context, r *http.Request, level slog.Level, body bool) {
	dumpFunc := httputil.DumpRequestOut
	if r.URL.Scheme == "" || r.URL.Host == "" {
		dumpFunc = httputil.DumpRequest
	}

	var binary bool

	if body && r.Body != nil && r.Body != http.NoBody {
		if body = isHumanReadable(r.Header.Get("Content-Type")); !body {
			binary = true
		}
	}

	b, err := dumpFunc(r, body)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP REQUEST", "error", err)
		return
	}

	if binary {
		buf := make([]byte, len(b)+len(holder))
		n := copy(buf, b)
		copy(buf[n:], holder)
		b = buf
	}

	slog.Log(ctx, level, "HTTP REQUEST", "dump", string(b))
}

// DumpHttpResponse logs the full HTTP response using slog at the specified log level.
// If body is true and the Content-Type is human-readable, the body is included in the dump.
// If the body is binary, a placeholder is appended instead.
func DumpHttpResponse(ctx context.Context, r *http.Response, level slog.Level, body bool) {
	var binary bool

	if body && r.Body != nil && r.Body != http.NoBody {
		if body = isHumanReadable(r.Header.Get("Content-Type")); !body {
			binary = true
		}
	}

	b, err := httputil.DumpResponse(r, body)
	if err != nil {
		slog.ErrorContext(ctx, "HTTP RESPONSE", "error", err)
		return
	}

	if binary {
		buf := make([]byte, len(b)+len(holder))
		n := copy(buf, b)
		copy(buf[n:], holder)
		b = buf
	}

	slog.Log(ctx, level, "HTTP RESPONSE", "dump", string(b))
}

// isHumanReadable reports whether the given Content-Type indicates text that is safe to log.
// Binary content (images, archives, octet-streams, multipart file uploads, audio/video, etc.) returns false.
func isHumanReadable(contentType string) bool {
	if contentType == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	mainType, subType, _ := strings.Cut(mediaType, "/")

	if mainType == "text" {
		return true
	}

	if mainType != "application" {
		return false
	}

	_, ok := types[subType]
	return ok
}

// Closer calls the Close method and logs any error via slog.
func Closer(ctx context.Context, c io.Closer) {
	if err := c.Close(); err != nil {
		slog.ErrorContext(ctx, "close resource", "error", err)
	}
}
