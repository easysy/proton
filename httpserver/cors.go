package httpserver

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions represents configuration for the CORS middleware.
type CORSOptions struct {
	AllowOrigins        []string          // List of allowed origins.
	AllowOriginsFunc    func(string) bool // Optional function for dynamic origin matching.
	AllowMethods        []string          // List of allowed HTTP methods.
	AllowHeaders        []string          // List of allowed request headers.
	ExposeHeaders       []string          // List of headers to expose to the client.
	MaxAge              int               // Max age (in seconds) to cache preflight.
	AllowCredentials    bool              // Whether credentials are allowed.
	AllowPrivateNetwork bool              // Whether requests from private networks are allowed.
}

func (c CORSOptions) checkOrigin(origin string) string {
	if origin == "" {
		return ""
	}

	var ok bool

	for _, v := range c.AllowOrigins {
		if origin == v {
			ok = true
		} else if v == "*" {
			if c.AllowCredentials {
				return origin
			}
			return "*"
		}
	}

	if ok {
		return origin
	}

	if c.AllowOriginsFunc != nil && c.AllowOriginsFunc(origin) {
		return origin
	}

	return ""
}

// AllowCORS sets headers for CORS mechanism supports secure.
func AllowCORS(opts *CORSOptions) func(next http.Handler) http.Handler {
	normalizeMethods(opts.AllowMethods)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin = opts.checkOrigin(origin); origin == "" {
				// Origin not allowed, respond with 403 on preflight or skip CORS headers
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)

			// Vary headers to ensure cache safety
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			// Allow credentials if configured
			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Preflight response
			if r.Method == http.MethodOptions {
				if len(opts.AllowHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ","))
				} else if requestHeaders := r.Header.Get("Access-Control-Request-Headers"); requestHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
				}

				w.Header().Set("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ","))

				if opts.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
				}

				if r.Header.Get("Private-Network") == "true" && opts.AllowPrivateNetwork {
					w.Header().Set("Access-Control-Allow-Private-Network", "true")
				}

				w.WriteHeader(http.StatusOK)
				return
			}

			if len(opts.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ","))
			}

			// Non-preflight request: continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

func normalizeMethods(methods []string) {
	for i := range methods {
		methods[i] = strings.ToUpper(methods[i])
	}
}
