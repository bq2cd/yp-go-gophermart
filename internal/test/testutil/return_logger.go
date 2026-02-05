package testutil

import (
	"context"
	"log/slog"
	"reflect"
	"slices"
)

const (
	returnLoggerSkipFrames           = 2
	returnLoggerFunctionLastname     = 2
	argsToKVPairsCapacityCoefficient = 2
)

type returnLoggerBase struct {
	level slog.Level
	args  []any
}

func (l *returnLoggerBase) log(retValues ...any) {
	fname := FunctionLastname(
		CallerFrame(returnLoggerSkipFrames),
		returnLoggerFunctionLastname,
	)
	args := slices.Concat(l.args, retValues)

	slog.Log(context.Background(), l.level, fname, argsToKVPairs(args...)...)
}

/////////////////////////////////////////////////////////////////////////////////

// ReturnLogger implements a convenience wrapper to log provided return value
// and return it without any modifications.
type ReturnLogger[T any] struct {
	returnLoggerBase
}

// NewReturnLogger creates an instance of [ReturnLogger].
func NewReturnLogger[T any](level slog.Level, args ...any) *ReturnLogger[T] {
	return &ReturnLogger[T]{
		returnLoggerBase: returnLoggerBase{
			level: level,
			args:  args,
		},
	}
}

// Log performs logging of a provided return value and returns it
// as is.
//
//nolint:ireturn
func (l *ReturnLogger[T]) Log(ret T) T {
	l.log(ret)

	return ret
}

/////////////////////////////////////////////////////////////////////////////////

// ReturnLogger2 implements a convenience wrapper to log provided return values (two values)
// and return them without any modifications.
type ReturnLogger2[A, B any] struct {
	returnLoggerBase
}

// NewReturnLogger2 creates an instance of [ReturnLogger2].
func NewReturnLogger2[A, B any](level slog.Level, args ...any) *ReturnLogger2[A, B] {
	return &ReturnLogger2[A, B]{
		returnLoggerBase: returnLoggerBase{
			level: level,
			args:  args,
		},
	}
}

// Log performs logging of a provided return values and returns them
// as is.
//
//nolint:ireturn
func (l *ReturnLogger2[A, B]) Log(ret1 A, ret2 B) (A, B) {
	l.log(ret1, ret2)

	return ret1, ret2
}

/////////////////////////////////////////////////////////////////////////////////

// ReturnLogger3 implements a convenience wrapper to log provided return values (three values)
// and return them without any modifications.
type ReturnLogger3[A, B, C any] struct {
	returnLoggerBase
}

// NewReturnLogger3 creates an instance of [ReturnLogger3].
func NewReturnLogger3[A, B, C any](level slog.Level, args ...any) *ReturnLogger3[A, B, C] {
	return &ReturnLogger3[A, B, C]{
		returnLoggerBase: returnLoggerBase{
			level: level,
			args:  args,
		},
	}
}

// Log performs logging of a provided return values and returns them
// as is.
//
//nolint:ireturn
func (l *ReturnLogger3[A, B, C]) Log(ret1 A, ret2 B, ret3 C) (A, B, C) {
	l.log(ret1, ret2, ret3)

	return ret1, ret2, ret3
}

/////////////////////////////////////////////////////////////////////////////////

func argsToKVPairs(args ...any) []any {
	kvPairs := make([]any, 0, len(args)*argsToKVPairsCapacityCoefficient)
	for _, arg := range args {
		t := reflect.TypeOf(arg)
		if t == nil {
			continue
		}

		kvPairs = append(kvPairs, t.String(), arg)
	}

	return kvPairs
}
