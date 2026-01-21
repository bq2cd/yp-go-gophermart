package testutil

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"go.uber.org/mock/gomock"
)

const (
	callerFrameMaxPC    = 16
	callerFrameBaseSkip = 2
)

// MockCtx is designed to match [context.Context] arguments
// when used with [gomock] mock calls.
//
//nolint:ireturn
func MockCtx() gomock.Matcher {
	ctx := reflect.TypeFor[context.Context]()

	return gomock.AssignableToTypeOf(ctx)
}

// CallerFrame will return [runtime.Frame] for its caller or
// a corresponding parent caller if [skip] is greater than zero.
func CallerFrame(skip int) runtime.Frame {
	pc := make([]uintptr, callerFrameMaxPC)
	n := runtime.Callers(skip+callerFrameBaseSkip, pc)
	frames := runtime.CallersFrames(pc[:n])
	first, _ := frames.Next()

	return first
}

// FunctionBasename returns the last portion after a dot of the fully-qualified function name.
func FunctionBasename(frame runtime.Frame) string {
	parts := strings.Split(frame.Function, ".")

	return parts[len(parts)-1]
}
