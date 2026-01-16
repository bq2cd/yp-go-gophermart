package service_test

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

const (
	exampleUserID = domain.UserID("user1")
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

func mockCtx() gomock.Matcher {
	ctx := reflect.TypeFor[context.Context]()

	return gomock.AssignableToTypeOf(ctx)
}
