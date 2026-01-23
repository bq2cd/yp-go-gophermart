package testutil

import (
	"context"
	"reflect"

	"go.uber.org/mock/gomock"
)

// MockCtx is designed to match [context.Context] arguments
// when used with [gomock] mock calls.
//
//nolint:ireturn
func MockCtx() gomock.Matcher {
	ctx := reflect.TypeFor[context.Context]()

	return gomock.AssignableToTypeOf(ctx)
}
