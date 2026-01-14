package service_test

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/mocks"
)

// Ensure [service.TokenManager] implements [handler.TokenService].
var _ handler.TokenService = (*service.TokenManager)(nil)

var _ = Describe("TokenManager", func() {
	var (
		ctrl              *gomock.Controller
		secretKeyProvider *mocks.MockSecretKeyProvider
		secretKeyBytes    []byte
		tokenLifetime     time.Duration
		tokenManager      *service.TokenManager
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		secretKeyProvider = mocks.NewMockSecretKeyProvider(ctrl)
		secretKeyBytes = []byte(`a secret key`)
		tokenLifetime = time.Minute
		tokenManager = service.NewTokenManager(secretKeyProvider, tokenLifetime)
	})

	Context("issuing a token", func() {
		var (
			userID     domain.UserID
			tokenValue domain.Token
			err        error
		)

		BeforeEach(func() {
			userID = domain.UserID("user1")
		})

		JustBeforeEach(func() {
			tokenValue, err = tokenManager.IssueToken(GinkgoT().Context(), userID)
		})

		When("secret key provider succeeds", func() {
			Context("returned secret key is non-empty", func() {
				BeforeEach(func() {
					secretKeyProvider.EXPECT().
						GetSecretKey(gomock.Any()).
						Return(secretKeyBytes, nil)
				})

				It("should generate token with correct claims", func() {
					Expect(err).NotTo(HaveOccurred())

					claims := jwt.RegisteredClaims{}
					token, err := jwt.ParseWithClaims(tokenValue.String(), &claims, func(_ *jwt.Token) (any, error) {
						return secretKeyBytes, nil
					})

					Expect(err).NotTo(HaveOccurred())

					Expect(token.Valid).To(BeTrue())
					Expect(token.Method.Alg()).To(Equal(jwt.SigningMethodHS256.Alg()))

					Expect(claims.Issuer).NotTo(BeEmpty())
					Expect(claims.Subject).To(Equal(string(userID)))
					Expect(claims.ExpiresAt.Time).To(BeTemporally("~", time.Now().Add(tokenLifetime), time.Second))
					Expect(claims.IssuedAt.Time).To(BeTemporally("<", time.Now()))
					Expect(claims.NotBefore.Time).To(BeTemporally("<", time.Now()))
				})
			})

			Context("returned secret key is empty", func() {
				BeforeEach(func() {
					secretKeyProvider.EXPECT().
						GetSecretKey(gomock.Any()).
						Return(nil, nil)
				})

				It("should generate token without errors", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(tokenValue.String()).NotTo(BeEmpty())
				})
			})
		})

		When("secret key provider fails", func() {
			BeforeEach(func() {
				secretKeyProvider.EXPECT().
					GetSecretKey(gomock.Any()).
					Return(nil, errors.New("oops"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("validating a token", func() {
		var (
			expectedUserID, actualUserID domain.UserID
			tokenValue                   domain.Token
			err                          error
			getTokenValueFn              func() (domain.Token, error)
			setupSecretKeyProviderFn     func()
		)

		BeforeEach(func() {
			expectedUserID = domain.UserID("user2")

			getTokenValueFn = func() (domain.Token, error) {
				secretKeyProvider.EXPECT().
					GetSecretKey(gomock.Any()).
					Return(secretKeyBytes, nil)

				return tokenManager.IssueToken(GinkgoT().Context(), expectedUserID)
			}

			setupSecretKeyProviderFn = func() {}
		})

		JustBeforeEach(func() {
			tokenValue, err = getTokenValueFn()
			Expect(err).NotTo(HaveOccurred())

			setupSecretKeyProviderFn()

			actualUserID, err = tokenManager.ValidateToken(GinkgoT().Context(), tokenValue)
		})

		When("secret key provider succeeds", func() {
			BeforeEach(func() {
				setupSecretKeyProviderFn = func() {
					secretKeyProvider.EXPECT().
						GetSecretKey(gomock.Any()).
						Return(secretKeyBytes, nil)
				}
			})

			Context("token contains non-empty issuer", func() {
				It("should return correct user ID without errors", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(actualUserID).To(Equal(expectedUserID))
				})
			})

			Context("token contains empty issuer", func() {
				BeforeEach(func() {
					expectedUserID = domain.UserIDEmptyValue
				})

				It("should return empty user ID with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(actualUserID).To(Equal(domain.UserIDEmptyValue))
				})
			})

			Context("arbitrary token with correct algorithm", func() {
				BeforeEach(func() {
					getTokenValueFn = func() (domain.Token, error) {
						token := jwt.New(jwt.SigningMethodHS256)
						tokenStr, err := token.SignedString(secretKeyBytes)

						return domain.Token(tokenStr), err
					}
				})

				It("should return empty user ID with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(actualUserID).To(Equal(domain.UserIDEmptyValue))
				})
			})

			Context("arbitrary token with incorrect algorithm", func() {
				BeforeEach(func() {
					getTokenValueFn = func() (domain.Token, error) {
						token := jwt.New(jwt.SigningMethodHS512)
						tokenStr, err := token.SignedString(secretKeyBytes)

						return domain.Token(tokenStr), err
					}
					setupSecretKeyProviderFn = func() {}
				})

				It("should return empty user ID with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(actualUserID).To(Equal(domain.UserIDEmptyValue))
				})
			})

			Context("forged token with different secret key", func() {
				BeforeEach(func() {
					getTokenValueFn = func() (domain.Token, error) {
						token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
							Issuer:    "GopherMart",
							Subject:   "user3",
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)),
						})
						tokenStr, err := token.SignedString([]byte(`forged key`))

						return domain.Token(tokenStr), err
					}
				})

				It("should return empty user ID with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(actualUserID).To(Equal(domain.UserIDEmptyValue))
				})
			})
		})

		When("secret key provider fails", func() {
			BeforeEach(func() {
				setupSecretKeyProviderFn = func() {
					secretKeyProvider.EXPECT().
						GetSecretKey(gomock.Any()).
						Return(nil, errors.New("oops"))
				}
			})

			It("should return empty user ID with error", func() {
				Expect(err).To(HaveOccurred())
				Expect(actualUserID).To(Equal(domain.UserIDEmptyValue))
			})
		})
	})
})
