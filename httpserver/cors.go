package httpserver

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions represents configuration for the CORS middleware.
type CORSOptions struct {
	// AllowOriginsFunc is an optional callback to dynamically allow or reject an origin.
	// If provided, it takes precedence over AllowOrigins.
	AllowOriginsFunc func(string) bool

	// AllowOrigins is the list of allowed origins.
	// An empty slice or a list containing "*" means all origins are allowed.
	AllowOrigins []string

	// AllowMethods is the list of allowed HTTP methods.
	// An empty slice or a list containing "*" means all methods are allowed.
	AllowMethods []string

	// AllowHeaders is the list of allowed request headers.
	// An empty slice or a list containing "*" means all headers are allowed.
	AllowHeaders []string

	// ExposeHeaders are the response headers that browsers are allowed to access
	// via client-side JavaScript (using XMLHttpRequest or Fetch).
	ExposeHeaders []string

	// MaxAge is the number of seconds a preflight request can be cached by the client.
	MaxAge int

	// AllowCredentials indicates whether cookies, authorization headers, and TLS
	// client certificates are exposed to the browser.
	AllowCredentials bool

	// AllowPrivateNetwork indicates whether requests from private network contexts
	// are allowed (controlled via the "Private-Network" CORS extension header).
	AllowPrivateNetwork bool

	// SkipStrictOriginCheck controls how requests without a valid Origin header are handled.
	//
	// By default, (false), requests missing an allowed Origin are rejected, ensuring strict
	// cross-origin enforcement for both preflight (OPTIONS) and non-preflight requests.
	//
	// If set to true, non-preflight requests (e.g., GET, POST) that do not include an
	// Origin header are allowed to pass through to the next handler. This is useful for:
	//
	//   - Same-origin requests (browsers often omit the Origin header)
	//   - Non-browser clients (curl, Go http.Client, etc.) that do not set Origin
	//
	// Security note: Setting this to true means requests without Origin are trusted.
	// You should only enable it if your API must support same-origin or non-browser clients.
	SkipStrictOriginCheck bool
}

type cors struct {
	isOriginAllowed         func(string) string
	allowedMethods          string
	isMethodAllowed         func(string) bool
	allowedHeaders          string
	isHeadersAllowed        func(string) bool
	exposeHeaders           string
	maxAge                  int
	isCredentialsAllowed    bool
	isPrivateNetworkAllowed bool
	skipStrictOriginCheck   bool
}

func makeAllowed(allowed []string, c func(string) string) map[string]struct{} {
	m := make(map[string]struct{}, len(allowed))
	for i := range allowed {
		allowed[i] = c(strings.TrimSpace(allowed[i]))
		m[allowed[i]] = struct{}{}
	}
	return m
}

func allowAll(string) bool {
	return true
}

func newCORS(opts *CORSOptions) *cors {
	if opts == nil {
		opts = &CORSOptions{}
	}

	for i := range opts.ExposeHeaders {
		opts.ExposeHeaders[i] = http.CanonicalHeaderKey(opts.ExposeHeaders[i])
	}

	c := &cors{
		isMethodAllowed:         allowAll,
		isHeadersAllowed:        allowAll,
		exposeHeaders:           strings.Join(opts.ExposeHeaders, ", "),
		maxAge:                  opts.MaxAge,
		isCredentialsAllowed:    opts.AllowCredentials,
		isPrivateNetworkAllowed: opts.AllowPrivateNetwork,
		skipStrictOriginCheck:   opts.SkipStrictOriginCheck,
	}

	allowOrigins := makeAllowed(opts.AllowOrigins, func(s string) string { return s })
	if _, all := allowOrigins["*"]; all {
		c.isOriginAllowed = func(origin string) string {
			if c.isCredentialsAllowed {
				return origin
			}
			return "*"
		}
	} else {
		c.isOriginAllowed = func(origin string) string {
			if opts.AllowOriginsFunc != nil && opts.AllowOriginsFunc(origin) {
				return origin
			}
			if _, is := allowOrigins[origin]; is {
				return origin
			}
			return ""
		}
	}

	allowMethods := makeAllowed(opts.AllowMethods, strings.ToUpper)
	if _, ok := allowMethods["*"]; !ok && len(allowMethods) != 0 {
		c.allowedMethods = strings.Join(opts.AllowMethods, ", ")
		c.isMethodAllowed = func(method string) bool {
			_, is := allowMethods[strings.ToUpper(method)]
			return is
		}
	}

	allowHeaders := makeAllowed(opts.AllowHeaders, http.CanonicalHeaderKey)
	if _, ok := allowHeaders["*"]; !ok && len(allowHeaders) != 0 {
		c.allowedHeaders = strings.Join(opts.AllowHeaders, ", ")
		c.isHeadersAllowed = func(headers string) bool {
			for _, header := range strings.Split(headers, ",") {
				if _, is := allowHeaders[http.CanonicalHeaderKey(strings.TrimSpace(header))]; !is {
					return false
				}
			}
			return true
		}
	}

	return c
}

func (c *cors) isAllowed(r *http.Request) bool {
	return c.isMethodAllowed(r.Header.Get("Access-Control-Request-Method")) && c.isHeadersAllowed(r.Header.Get("Access-Control-Request-Headers"))
}

// AllowCORS sets headers for CORS mechanism supports secure.
func AllowCORS(opts *CORSOptions) func(next http.Handler) http.Handler {
	c := newCORS(opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			originHeader := r.Header.Get("Origin")

			origin := c.isOriginAllowed(originHeader)
			if origin == "" {
				if r.Method != http.MethodOptions && c.skipStrictOriginCheck && originHeader == "" {
					next.ServeHTTP(w, r)
				}

				return
			}

			if r.Method == http.MethodOptions && !c.isAllowed(r) {
				return
			}

			// Vary headers to ensure cache safety
			w.Header().Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")

			w.Header().Set("Access-Control-Allow-Origin", origin)

			// Allow credentials if configured
			if c.isCredentialsAllowed {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Preflight response
			if r.Method == http.MethodOptions {
				if c.allowedMethods != "" {
					w.Header().Set("Access-Control-Allow-Methods", c.allowedMethods)
				} else if requestMethod := r.Header.Get("Access-Control-Request-Method"); requestMethod != "" {
					w.Header().Set("Access-Control-Allow-Methods", strings.ToUpper(requestMethod))
				}

				if c.allowedHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", c.allowedHeaders)
				} else if requestHeaders := r.Header.Get("Access-Control-Request-Headers"); requestHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
				}

				if c.maxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.maxAge))
				}

				if strings.EqualFold(r.Header.Get("Private-Network"), "true") && c.isPrivateNetworkAllowed {
					w.Header().Set("Access-Control-Allow-Private-Network", "true")
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			if c.exposeHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", c.exposeHeaders)
			}

			// Non-preflight request: continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
