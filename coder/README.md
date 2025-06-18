# coder

### The `coder` package implements three interfaces with debug logging:

- *Encoder* encodes and writes values to an output stream.
- *Decoder* reads and decodes values from an input stream.
- *Coder* is a pair of Encoder and Decoder.

## Getting Started

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/easysy/proton/coder"
)

func main() {
	ctx := context.Background()

	cdrJSON := coder.NewCoder("application/json", json.Marshal, json.Unmarshal)

	var buf bytes.Buffer

	in := &struct {
		A string `json:"a"`
	}{A: "AAA"}

	if err := cdrJSON.Encode(ctx, &buf, in); err != nil {
		panic(err)
	}

	fmt.Printf("encoded: %s\n", buf.String())
	// encoded: {"a":"AAA"}

	out := &struct {
		A string `json:"a"`
	}{}

	if err := cdrJSON.Decode(ctx, &buf, out); err != nil {
		panic(err)
	}

	fmt.Printf("decoded: %+v\n", out)
	// decoded: &{A:AAA}
}

```