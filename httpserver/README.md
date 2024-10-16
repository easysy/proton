# httpserver

### The `httpserver` package implements the core functionality of a [coder](https://github.com/easysy/proton/blob/main/coder/README.md)-based http server.

## Getting Started

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/easysy/proton/coder"
	"github.com/easysy/proton/httpserver"
)

func main() {
	cdrJSON := coder.NewCoder("application/json", json.Marshal, json.Unmarshal, false)

	fmtJSON := httpserver.NewFormatter(cdrJSON)

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodGet {
			req := &struct {
				// some fields
			}{}

			if err := fmtJSON.Decode(ctx, r.Body, req); err != nil {
				panic(err)
			}
		}

		res := &struct {
			ID int `json:"id"`
		}{ID: 1}

		fmtJSON.WriteResponse(ctx, w, http.StatusOK, res)
	})

	http.Handle("/example/", handlerFunc)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

```

### For all responses with a body:

- The default is to use the `Content-Type` set in
  the [coder](https://github.com/easysy/proton/blob/main/coder/README.md).
- If you don't set the `Content-Type` in
  the [coder](https://github.com/easysy/proton/blob/main/coder/README.md), it
  will be set automatically by the [net/http](https://pkg.go.dev/net/http) package.
- If you need to set a different `Content-Type` you must set it before calling `WriteResponse`.

### For all responses without a body:

- `Content-Type` will not be set by default.
- If you need to set `Content-Type` you must set it before calling `WriteResponse`.

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/easysy/proton/httpserver"
)

func main() {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Hello!")
	})

	handler := http.NewServeMux()
	handler.HandleFunc("/example/", handlerFunc)

	srv := new(http.Server)
	srv.Addr = ":8080"
	srv.Handler = handler

	hcr := new(httpserver.Controller)
	hcr.Server = srv
	hcr.GracefulTimeout = time.Second * 10

	if err := hcr.Start(); err != nil {
		panic(err)
	}
}

```

### The `httpserver` package contains functions that are used as middleware on the http server side.

## Getting Started

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/easysy/proton/httpserver"
)

func main() {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "Hello World!"); err != nil {
			panic(err)
		}
	})

	http.Handle("/example/", handlerFunc)

	corsOpts := new(httpserver.CORSOptions)

	corsOpts.AllowOrigins = []string{"*"}
	corsOpts.AllowMethods = []string{"OPTIONS", "GET", "POST", "PUT", "PATCH", "DELETE"}
	corsOpts.AllowHeaders = []string{"Authorization", "Content-Type"}
	corsOpts.MaxAge = 86400
	corsOpts.AllowCredentials = false

	handler := httpserver.MiddlewareSequencer(
		http.DefaultServeMux,
		httpserver.DumpHttp(slog.LevelDebug),
		httpserver.Timer(slog.LevelInfo),
		httpserver.Tracer,
		httpserver.AllowCORS(corsOpts),
		httpserver.PanicCatcher,
	)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}

```
