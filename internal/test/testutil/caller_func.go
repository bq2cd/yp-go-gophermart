package testutil

import (
	"runtime"
	"strings"
)

const (
	callerFrameMaxPC    = 16
	callerFrameBaseSkip = 2
)

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

// FunctionLastname returns the last N parts after a dot of the fully-qualified function name.
func FunctionLastname(frame runtime.Frame, last int) string {
	parts := strings.Split(frame.Function, ".")
	plen := len(parts)

	if plen > last {
		parts = parts[plen-last : plen]
	}

	return strings.Join(parts, ".")
}
