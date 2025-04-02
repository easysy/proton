package httpclient

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/easysy/proton/coder"
)

type Client interface {
	coder.Coder
	Request(ctx context.Context, method, url string, body any, f func(*http.Request)) (*http.Response, error)
	SendFile(ctx context.Context, url, key, filename string, body io.ReadSeeker, f func(*http.Request)) (*http.Response, error)
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
// To add additional data to the request, use the optional function f (e.g., for adding headers).
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

// SendFile sends a file as a multipart form upload via an HTTP POST request based on the given URL, key, name and body.
// To add additional data to the request, use the optional function f (e.g., for adding headers).
//   - key: the form field name for the file;
//   - name: the name of the file being uploaded.
func (c *protoClient) SendFile(ctx context.Context, url, key, name string, body io.ReadSeeker, f func(*http.Request)) (*http.Response, error) {
	// Create a buffer for the multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a form file field
	part, err := writer.CreateFormFile(key, name)
	if err != nil {
		return nil, err
	}

	// Copy the file content into the form
	if _, err = io.Copy(part, body); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequestWithContext(ctx, http.MethodPost, url, &buf); err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	if f != nil {
		f(request)
	}

	return c.Do(request)
}
