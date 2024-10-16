package coder

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

const ContentType = "Content-Type"

// An Encoder encodes and writes values to an output stream.
type Encoder interface {
	Encode(ctx context.Context, w io.Writer, v any) error
}

type encoder struct {
	f   func(v any) ([]byte, error)
	raw bool
}

// NewEncoder returns a new Encoder that writes to w.
// If 'raw' is true, the debug log will print raw bytes.
func NewEncoder(marshal func(v any) ([]byte, error), raw bool) Encoder {
	return &encoder{f: marshal, raw: raw}
}

// Encode encodes the value pointed to by v and writes it to the stream.
// It will panic if encoder function not set.
func (e *encoder) Encode(ctx context.Context, w io.Writer, v any) error {
	enabled := slog.Default().Enabled(ctx, slog.LevelDebug)

	if enabled {
		slog.DebugContext(ctx, "encoder input", "value", v)
	}

	p, err := e.f(v)
	if err != nil {
		return err
	}

	if enabled {
		var attr slog.Attr
		if e.raw {
			attr = slog.String("bytes", fmt.Sprintf("% x", p))
		} else {
			attr = slog.String("value", string(p))
		}
		slog.DebugContext(ctx, "encoder output", attr, "len", len(p))
	}

	if _, err = w.Write(p); err != nil {
		return err
	}

	return nil
}

// A Decoder reads and decodes values from an input stream.
type Decoder interface {
	Decode(ctx context.Context, r io.Reader, v any) error
}

type decoder struct {
	f   func(data []byte, v any) error
	raw bool
}

// NewDecoder returns a new Decoder that reads from r.
// If 'raw' is true, the debug log will print raw bytes.
func NewDecoder(unmarshal func(data []byte, v any) error, raw bool) Decoder {
	return &decoder{f: unmarshal, raw: raw}
}

// Decode reads the next encoded value from its input and stores it in the value pointed to by v.
// It will panic if decoder function not set.
func (d *decoder) Decode(ctx context.Context, r io.Reader, v any) error {
	p, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	enabled := slog.Default().Enabled(ctx, slog.LevelDebug)

	if enabled {
		var attr slog.Attr
		if d.raw {
			attr = slog.String("bytes", fmt.Sprintf("% x", p))
		} else {
			attr = slog.String("value", string(p))
		}
		slog.DebugContext(ctx, "decoder input", attr, "len", len(p))
	}

	if err = d.f(p, v); err != nil {
		return err
	}

	if enabled {
		slog.DebugContext(ctx, "decoder output", "value", v)
	}

	return nil
}

// A Coder is a pair of Encoder and Decoder.
type Coder interface {
	ContentType() string
	Encoder
	Decoder
}

type coder struct {
	t string
	Encoder
	Decoder
}

// NewCoder returns a new Coder.
// If 'raw' is true, the debug log will print raw bytes.
func NewCoder(contentType string, marshal func(v any) ([]byte, error), unmarshal func(data []byte, v any) error, raw bool) Coder {
	return &coder{t: contentType, Encoder: NewEncoder(marshal, raw), Decoder: NewDecoder(unmarshal, raw)}
}

// ContentType returns a string value representing the Coder type.
// Use as the ContentType header of HTTP requests.
func (c coder) ContentType() string {
	return c.t
}
