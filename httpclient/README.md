# httpclient

### The `httpclient` package implements the core functionality of a [coder](https://github.com/easysy/proton/blob/main/coder/README.md)-based http client.

## Getting Started

```go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/easysy/proton/coder"
	"github.com/easysy/proton/httpclient"
	"github.com/easysy/proton/log"
)

func main() {
	cdrJSON := coder.NewCoder("application/json", json.Marshal, json.Unmarshal)

	clientJSON := httpclient.New(cdrJSON, http.DefaultClient)

	URL := "http://localhost:8080/example/"

	params := make(url.Values)
	params.Add("id", "1")

	// To add additional data to the request, use the optional function f(*http.Request)
	f := func(r *http.Request) {
		r.Header.Set("Accept", "application/json")
		r.URL.RawQuery = params.Encode()
	}

	ctx := context.Background()

	resp, err := clientJSON.Request(ctx, http.MethodGet, URL, nil, f)
	if err != nil {
		panic(err)
	}

	defer log.Closer(ctx, resp.Body)

	res := &struct {
		// some fields
	}{}

	if err = clientJSON.Decode(ctx, resp.Body, res); err != nil {
		panic(err)
	}
}

```

```go
package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/easysy/proton/coder"
	"github.com/easysy/proton/httpclient"
	"github.com/easysy/proton/log"
)

func main() {
	cdrJSON := coder.NewCoder("application/json", json.Marshal, json.Unmarshal)

	clientJSON := httpclient.New(cdrJSON, http.DefaultClient)

	URL := "http://localhost:8080/v1/example/"

	req := &struct {
		ID int `json:"id"`
	}{ID: 1}

	ctx := context.Background()

	resp, err := clientJSON.Request(ctx, http.MethodPost, URL, req, nil)
	if err != nil {
		panic(err)
	}

	defer log.Closer(ctx, resp.Body)

	res := &struct {
		// some fields
	}{}

	if err = clientJSON.Decode(ctx, resp.Body, res); err != nil {
		panic(err)
	}
}

```

### The `httpclient` package contains functions that are used as middleware on the http client side.

## Getting Started

```go
package main

import (
	"log/slog"
	"net/http"

	"github.com/easysy/proton/httpclient"
)

func main() {
	transport := httpclient.RoundTripperSequencer(
		http.DefaultTransport,
		httpclient.DumpHttp(slog.LevelDebug, true),
		httpclient.Timer(slog.LevelInfo),
		httpclient.Tracer,
		httpclient.PanicCatcher,
	)

	hct := new(http.Client)
	hct.Transport = transport
}

```