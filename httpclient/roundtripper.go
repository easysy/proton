package httpclient

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/easysy/proton/log"
	"github.com/easysy/proton/sgen"
)

func init() {
	id.Configure("", sgen.UpLetters.Append(sgen.LowLetters, sgen.Nums), 12)
}

var id sgen.RandomString

// The RoundTripper type is an adapter to allow the use of ordinary functions as HTTP round trippers.
// If f is a function with the appropriate signature, Func(f) is a RoundTripper that calls f.
type RoundTripper func(*http.Request) (*http.Response, error)

// RoundTrip calls f(r).
func (f RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// RoundTripperSequencer chains http.RoundTrippers in a chain.
func RoundTripperSequencer(baseRoundTripper http.RoundTripper, rts ...func(http.RoundTripper) http.RoundTripper) http.RoundTripper {
	for _, f := range rts {
		baseRoundTripper = f(baseRoundTripper)
	}
	return baseRoundTripper
}

// Tracer adds trace ID to the request context.
func Tracer(next http.RoundTripper) http.RoundTripper {
	return RoundTripper(func(r *http.Request) (*http.Response, error) {
		ctx := context.WithValue(r.Context(), log.TraceCtxKey, id.Generate())
		return next.RoundTrip(r.Clone(ctx))
	})
}

// Timer measures the time taken by http.RoundTripper.
func Timer(level slog.Level) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripper(func(r *http.Request) (*http.Response, error) {
			ctx := r.Context()
			if slog.Default().Enabled(ctx, level) {
				defer func(start time.Time) {
					slog.Log(ctx, level, "finished",
						slog.Group("request",
							slog.String("method", r.Method),
							slog.String("url", r.URL.String()),
						),
						slog.String("duration", time.Since(start).String()))
				}(time.Now())
			}
			return next.RoundTrip(r)
		})
	}
}

// PanicCatcher handles panics in http.RoundTripper.
func PanicCatcher(next http.RoundTripper) http.RoundTripper {
	return RoundTripper(func(r *http.Request) (*http.Response, error) {
		defer func() {
			if rec := recover(); rec != nil {
				if slog.Default().Enabled(r.Context(), slog.LevelError) {
					slog.Error("panic", "recover", rec, "stack", debug.Stack())
				}
			}
		}()
		return next.RoundTrip(r)
	})
}

// DumpHttp dumps the HTTP request and response, and prints out with logFunc.
func DumpHttp(level slog.Level, body bool) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripper(func(r *http.Request) (*http.Response, error) {
			ctx := r.Context()
			if slog.Default().Enabled(ctx, level) {
				log.DumpHttpRequest(ctx, r, level, body)

				response, err := next.RoundTrip(r)
				if err != nil {
					return nil, err
				}

				log.DumpHttpResponse(ctx, response, level, body)

				return response, nil
			}
			return next.RoundTrip(r)
		})
	}
}
