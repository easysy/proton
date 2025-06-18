package httpserver_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/easysy/proton/coder"
	"github.com/easysy/proton/httpserver"
)

func equal(t *testing.T, exp, got any) {
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("Not equal:\nexp: %v\ngot: %v", exp, got)
	}
}

var (
	cdrJSON = coder.NewCoder("application/json", json.Marshal, json.Unmarshal)
	cdr     = coder.NewCoder("", json.Marshal, json.Unmarshal)
)

type serverTestStruct struct {
	Field int
}

type clientTestStruct struct {
	Field string
}

func makeRequestWriteResponse(t *testing.T, method string, srv *httptest.Server, in any, out any) {
	var reader io.Reader
	if in != nil {
		buf := new(bytes.Buffer)

		err := json.NewEncoder(buf).Encode(in)
		equal(t, nil, err)

		reader = buf
	}

	req, err := http.NewRequest(method, srv.URL+"/path", reader)
	equal(t, nil, err)

	var resp *http.Response
	resp, err = srv.Client().Do(req)
	equal(t, nil, err)

	defer func() { _ = resp.Body.Close() }()

	output := &clientTestStruct{}

	err = json.NewDecoder(resp.Body).Decode(output)
	equal(t, nil, err)
	equal(t, out, output)
}

func TestProtoFormatter_WriteResponse(t *testing.T) {
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
			fmtJSON := httpserver.NewFormatter(cdrJSON)

			srv := httptest.NewServer(httpserver.DumpHttp(slog.LevelDebug, 1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				equal(t, r.URL.String(), "/path")

				ctx := r.Context()

				if test.input != nil {
					input := &serverTestStruct{}

					err := fmtJSON.Decode(ctx, r.Body, input)
					equal(t, nil, err)
					equal(t, test.input, input)
				}

				fmtJSON.WriteResponse(ctx, w, http.StatusOK, test.output)
			})))
			defer srv.Close()

			makeRequestWriteResponse(t, test.method, srv, test.input, test.output)
		})
	}
}

func makeRequestContentType(t *testing.T, srv *httptest.Server, expContentType string) {
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	equal(t, nil, err)

	var resp *http.Response
	resp, err = srv.Client().Do(req)
	equal(t, nil, err)
	equal(t, expContentType, resp.Header.Get(coder.ContentType))
}

func TestProtoFormatter_ResponseContentType(t *testing.T) {
	output := &clientTestStruct{Field: "example"}
	var tests = []struct {
		name              string
		coder             coder.Coder
		customContentType string
		expContentType    string
		output            any
	}{
		{
			name:  "empty content type",
			coder: cdrJSON,
		},
		{
			name:           "coder content type",
			coder:          cdrJSON,
			expContentType: "application/json",
			output:         output,
		},
		{
			name:              "custom content type",
			coder:             cdrJSON,
			customContentType: "application/xml",
			expContentType:    "application/xml",
			output:            output,
		},
		{
			name:           "default content type",
			coder:          cdr,
			expContentType: "text/plain; charset=utf-8",
			output:         output,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fmtJSON := httpserver.NewFormatter(test.coder)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				equal(t, r.URL.String(), "/path")

				if test.customContentType != "" {
					w.Header().Set(coder.ContentType, test.customContentType)
				}

				fmtJSON.WriteResponse(r.Context(), w, http.StatusOK, test.output)
			}))
			defer srv.Close()

			makeRequestContentType(t, srv, test.expContentType)
		})
	}
}
