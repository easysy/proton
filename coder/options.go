package coder

import "log/slog"

type Options interface {
	applyEnc(*encoder)
	applyDec(*decoder)
}

type raw struct{}

func (r raw) applyEnc(e *encoder) {
	e.raw = true
}

func (r raw) applyDec(d *decoder) {
	d.raw = true
}

func WithRawBytesLogging() Options {
	return raw{}
}

type level slog.Level

func (l level) applyEnc(e *encoder) {
	e.lvl = slog.Level(l)
}

func (l level) applyDec(d *decoder) {
	d.lvl = slog.Level(l)
}

func WithLogLevel(lvl slog.Level) Options {
	return level(lvl)
}
