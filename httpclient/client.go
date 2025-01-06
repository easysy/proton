package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/easysy/proton/coder"
)

type Client interface {
	coder.Coder
	Request(ctx context.Context, method, url string, body any, f func(*http.Request)) (*http.Response, error)
}

type protoClient struct {
	coder.Coder
	*http.Client
}

// New returns a new Client.
func New(coder coder.Coder, client *http.Client) Client {
	return &protoClient{Coder: coder, Client: client}
}

// Request sends an HTTP request based on the given method, URL, and optional body, and returns an HTTP response.
// To add additional data to the request, use the optional function f.
func (c *protoClient) Request(ctx context.Context, method, url string, body any, f func(*http.Request)) (*http.Response, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := c.Encode(ctx, buf, body); err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, err
	}

	if buf != nil && c.ContentType() != "" {
		request.Header.Set(coder.ContentType, c.ContentType())
	}

	if f != nil {
		f(request)
	}

	return c.Do(request)
}
