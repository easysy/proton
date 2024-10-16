package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"time"

	"github.com/easysy/proton/utils/log"
	"github.com/easysy/proton/utils/sgen"
)

func init() {
	id.Configure("", sgen.UpLetters.Append(sgen.LowLetters, sgen.Nums), 12)
}

var id sgen.RandomString

// MiddlewareSequencer chains middleware functions in a chain.
func MiddlewareSequencer(baseHandler http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for _, f := range mws {
		baseHandler = f(baseHandler)
	}
	return baseHandler
}

// Tracer adds trace ID to the request context.
func Tracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), log.TraceCtxKey, id.Generate())
		next.ServeHTTP(w, r.Clone(ctx))
	})
}

// Timer measures the time taken by http.HandlerFunc.
func Timer(level slog.Level) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if slog.Default().Enabled(ctx, level) {
				defer func(start time.Time) {
					slog.Log(ctx, level, "finished",
						slog.Group("request",
							slog.String("method", r.Method),
							slog.String("url", r.RequestURI),
						),
						slog.String("duration", time.Since(start).String()),
					)
				}(time.Now())
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PanicCatcher handles panics in http.HandlerFunc.
func PanicCatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				if slog.Default().Enabled(r.Context(), slog.LevelError) {
					slog.Error("panic", "recover", rec, "stack", debug.Stack())
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// DumpHttp dumps the HTTP request and response, and prints out.
func DumpHttp(level slog.Level) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if slog.Default().Enabled(ctx, level) {
				log.DumpHttpRequest(ctx, r, level)

				recorder := httptest.NewRecorder()

				next.ServeHTTP(recorder, r)

				for key, values := range recorder.Header() {
					w.Header().Del(key)
					for _, value := range values {
						w.Header().Set(key, value)
					}
				}

				w.WriteHeader(recorder.Code)

				response := recorder.Result()
				response.ContentLength, _ = recorder.Body.WriteTo(w)

				log.DumpHttpResponse(ctx, response, level)

				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
