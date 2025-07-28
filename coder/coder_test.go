package coder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/easysy/proton/coder"
)

func equal(t *testing.T, exp, got any) {
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("Not equal:\nexp: %v\ngot: %v", exp, got)
	}
}

type testStruct struct {
	Field string `json:"field"`
}

func TestEncoder_Encode(t *testing.T) {
	encoder := coder.NewEncoder(json.Marshal)
	var tests = []struct {
		name   string
		input  *testStruct
		output []byte
		err    error
	}{
		{
			name:   "successful encode",
			input:  &testStruct{Field: "example"},
			output: []byte("{\"field\":\"example\"}"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := encoder.Encode(context.Background(), w, test.input)
			if test.err != nil {
				equal(t, test.err.Error(), err.Error())
			} else {
				equal(t, nil, err)
				equal(t, test.output, w.Bytes())
			}
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	decoder := coder.NewDecoder(json.Unmarshal)
	var tests = []struct {
		name   string
		input  []byte
		output *testStruct
		err    error
	}{
		{
			name:   "successful decode",
			input:  []byte("{\"field\":\"example\"}"),
			output: &testStruct{Field: "example"},
		},
		{
			name:  "unexpected end of JSON input",
			input: []byte("{\"field\":\"example\""),
			err:   errors.New("unexpected end of JSON input"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &bytes.Buffer{}
			r.Write(test.input)

			v := new(testStruct)

			err := decoder.Decode(context.Background(), r, v)
			if test.err != nil {
				equal(t, test.err.Error(), err.Error())
			} else {
				equal(t, nil, err)
				equal(t, test.output, v)
			}
		})
	}
}
