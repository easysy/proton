package httpclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/easysy/proton/coder"
	"github.com/easysy/proton/httpclient"
)

func equal(t *testing.T, exp, got any) {
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("Not equal:\nexp: %v\ngot: %v", exp, got)
	}
}

var cdrJSON = coder.NewCoder("application/json", json.Marshal, json.Unmarshal, false)

type serverTestStruct struct {
	Field int
}

type clientTestStruct struct {
	Field string
}

func makeTestSrvRequest(t *testing.T, in, out any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equal(t, r.URL.String(), "/path")

		if in != nil {
			input := &serverTestStruct{}

			err := json.NewDecoder(r.Body).Decode(input)
			equal(t, nil, err)
			equal(t, in, input)
		}

		err := json.NewEncoder(w).Encode(out)
		equal(t, nil, err)
	}))
}

func TestProtoClient_Request(t *testing.T) {
	var tests = []struct {
		name   string
		method string
		input  any
		output any
		err    error
	}{
		{
			name:   "successful GET request",
			method: http.MethodGet,
			output: &clientTestStruct{Field: "example"},
		},
		{
			name:   "successful POST request",
			method: http.MethodPost,
			input:  &serverTestStruct{Field: 1},
			output: &clientTestStruct{Field: "example"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := makeTestSrvRequest(t, test.input, test.output)
			defer srv.Close()

			clt := httpclient.New(cdrJSON, srv.Client())

			ctx := context.Background()

			resp, err := clt.Request(ctx, test.method, srv.URL+"/path", test.input, nil)
			equal(t, nil, err)

			defer func() { _ = resp.Body.Close() }()

			output := &clientTestStruct{}

			err = clt.Decode(ctx, resp.Body, output)
			equal(t, nil, err)
			equal(t, test.output, output)
		})
	}
}

func makeTestSrvContentType(t *testing.T, expContentType string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equal(t, r.URL.String(), "/path")
		equal(t, expContentType, r.Header.Get(coder.ContentType))
	}))
}

func TestProtoClient_ContentType(t *testing.T) {
	var tests = []struct {
		name        string
		method      string
		coder       coder.Coder
		contentType string
		input       any
		rFunc       func(r *http.Request)
	}{
		{
			name:   "GET request default",
			method: http.MethodGet,
		},
		{
			name:        "GET request with coder content type",
			method:      http.MethodGet,
			contentType: cdrJSON.ContentType(),
			rFunc: func(r *http.Request) {
				r.Header.Set(coder.ContentType, cdrJSON.ContentType())
			},
		},
		{
			name:        "GET request with custom content type",
			method:      http.MethodGet,
			contentType: "application/xml",
			rFunc: func(r *http.Request) {
				r.Header.Set(coder.ContentType, "application/xml")
			},
		},
		{
			name:        "POST request with coder content type",
			contentType: cdrJSON.ContentType(),
			method:      http.MethodPost,
			input:       &serverTestStruct{Field: 1},
		},
		{
			name:        "POST request with custom content type",
			contentType: "application/xml",
			method:      http.MethodPost,
			input:       &serverTestStruct{Field: 1},
			rFunc: func(r *http.Request) {
				r.Header.Set(coder.ContentType, "application/xml")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := makeTestSrvContentType(t, test.contentType)
			defer srv.Close()

			clt := httpclient.New(cdrJSON, srv.Client())

			_, err := clt.Request(context.Background(), test.method, srv.URL+"/path", test.input, test.rFunc)
			equal(t, nil, err)
		})
	}
}
