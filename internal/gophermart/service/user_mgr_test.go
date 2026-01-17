package service_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/mocks"
)

// Ensure [service.UserManager] implements [handler.UserService].
var _ handler.UserService = (*service.UserManager)(nil)

var _ = Describe("UserManager", func() {
	var (
		userRepo      *mocks.MockUserRepository
		userMgr       *service.UserManager
		userID        domain.UserID
		passwordPlain domain.PasswordPlain
		err           error
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())

		userRepo = mocks.NewMockUserRepository(ctrl)
		userMgr = service.NewUserManager(userRepo)
		userID = domain.UserID("user1")
		passwordPlain = domain.PasswordPlain("example-password@123")
	})

	Describe("registering a user", func() {
		JustBeforeEach(func() {
			err = userMgr.Register(GinkgoT().Context(), userID, passwordPlain)
		})

		When("user provides empty password", func() {
			BeforeEach(func() {
				passwordPlain = domain.PasswordPlain("")
			})

			It("should return ErrEmptyPassword", func() {
				Expect(err).To(MatchError(domain.ErrEmptyPassword))
			})
		})

		Context("user repository is healthy", func() {
			When("user is brand new", func() {
				BeforeEach(func() {
					userRepo.EXPECT().
						CreateUser(gomock.Any(), userID, NewPasswordHashMatcher(userMgr, passwordPlain)).
						Return(true, nil)
				})

				It("should be created", func() {
					Expect(err).To(Succeed())
				})
			})

			When("user already exists", func() {
				BeforeEach(func() {
					userRepo.EXPECT().
						CreateUser(gomock.Any(), userID, NewPasswordHashMatcher(userMgr, passwordPlain)).
						Return(false, nil)
				})

				It("should return ErrUserIDConflict", func() {
					Expect(err).To(MatchError(domain.ErrUserIDConflict))
				})
			})
		})

		Context("user repository fails", func() {
			BeforeEach(func() {
				userRepo.EXPECT().
					CreateUser(gomock.Any(), userID, NewPasswordHashMatcher(userMgr, passwordPlain)).
					Return(false, errors.New("oops"))
			})

			It("should return error", func() {
				Expect(err).To(MatchError(ContainSubstring("oops")))
			})
		})
	})

	Describe("authenticating a user", func() {
		var (
			passwordHash domain.PasswordHash
		)

		BeforeEach(func() {
			passwordHash = generatePasswordHash(userMgr, passwordPlain)
		})

		JustBeforeEach(func() {
			err = userMgr.Authenticate(GinkgoT().Context(), userID, passwordPlain)
		})

		Context("user repository is healthy", func() {
			Context("user exists", func() {
				BeforeEach(func() {
					userRepo.EXPECT().
						GetPasswordHash(GinkgoT().Context(), userID).
						Return(passwordHash, nil)
				})

				When("password is valid", func() {
					It("should succeed", func() {
						Expect(err).To(Succeed())
					})
				})

				DescribeTableSubtree("password is incorrect",
					func(passwd string) {
						BeforeEach(func() {
							passwordPlain = domain.PasswordPlain(passwd)
						})

						It("should return ErrUserAuthenticationFailed", func() {
							Expect(err).To(MatchError(domain.ErrUserAuthenticationFailed))
						})
					},
					Entry("a random password", "a random password"),
					Entry("empty password", ""),
				)
			})

			Context("user does not exist", func() {
				BeforeEach(func() {
					userRepo.EXPECT().
						GetPasswordHash(GinkgoT().Context(), userID).
						Return(nil, nil)
				})

				It("should return ErrUserAuthenticationFailed", func() {
					Expect(err).To(MatchError(domain.ErrUserAuthenticationFailed))
				})
			})
		})

		Context("user repository fails", func() {
			BeforeEach(func() {
				userRepo.EXPECT().
					GetPasswordHash(gomock.Any(), userID).
					Return(nil, errors.New("oops"))
			})

			It("should return error", func() {
				Expect(err).To(MatchError(ContainSubstring("oops")))
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

type PasswordHashMatcher struct {
	userMgr       *service.UserManager
	passwordPlain domain.PasswordPlain
}

func NewPasswordHashMatcher(userMgr *service.UserManager, passwordPlain domain.PasswordPlain) gomock.Matcher {
	return &PasswordHashMatcher{
		userMgr:       userMgr,
		passwordPlain: passwordPlain,
	}
}

func (m *PasswordHashMatcher) Matches(value any) bool {
	passwordHash, ok := value.(domain.PasswordHash)
	if !ok {
		return false
	}

	err := m.userMgr.ValidatePasswordHash(m.passwordPlain, passwordHash)

	return err == nil
}

func (m *PasswordHashMatcher) String() string {
	return fmt.Sprintf("password hash matches plaintext password (%s)", m.passwordPlain)
}

func generatePasswordHash(userMgr *service.UserManager, passwordPlain domain.PasswordPlain) domain.PasswordHash {
	GinkgoHelper()

	passwordHash, err := userMgr.GeneratePasswordHash(passwordPlain)
	Expect(err).To(Succeed())

	return passwordHash
}
