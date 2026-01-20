package testutil

import (
	"reflect"
	"slices"

	"github.com/go-logr/logr"
)

const (
	returnLoggerSkipFrames           = 2
	argsToKVPairsCapacityCoefficient = 2
)

type returnLoggerBase struct {
	logger logr.Logger
	args   []any
}

func (l *returnLoggerBase) log(retValues ...any) {
	fname := FunctionBasename(CallerFrame(returnLoggerSkipFrames))
	args := slices.Concat(l.args, retValues)

	l.logger.Info(fname, argsToKVPairs(args...)...)
}

// ReturnLogger implements a convenience wrapper to log provided return value
// and return it without any modifications.
type ReturnLogger[T any] struct {
	returnLoggerBase
}

// NewReturnLogger creates an instance of [ReturnLogger].
func NewReturnLogger[T any](logger logr.Logger, args ...any) *ReturnLogger[T] {
	return &ReturnLogger[T]{
		returnLoggerBase: returnLoggerBase{
			logger: logger,
			args:   args,
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

// ReturnLogger2 implements a convenience wrapper to log provided return values (two values)
// and return them without any modifications.
type ReturnLogger2[A, B any] struct {
	returnLoggerBase
}

// NewReturnLogger2 creates an instance of [ReturnLogger2].
func NewReturnLogger2[A, B any](logger logr.Logger, args ...any) *ReturnLogger2[A, B] {
	return &ReturnLogger2[A, B]{
		returnLoggerBase: returnLoggerBase{
			logger: logger,
			args:   args,
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
