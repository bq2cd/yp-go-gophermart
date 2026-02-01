package sqldatabase_test

import (
	"crypto/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

const (
	exampleUserLogin            = "test-user-1"
	exampleUserNonExistentLogin = "non-existent-user"
)

var _ = Describe("StoreUser", func() {
	var storage *sqldatabase.Storage

	Context("sqlite database", func() {
		var query gorm.Interface[models.User]

		BeforeEach(OncePerOrdered, func() {
			tempStorage := createTempStorage()
			DeferCleanup(tempStorage.Cleanup)

			storage = tempStorage.Storage
		})

		BeforeEach(func() {
			query = sqldatabase.Query[models.User](storage)
		})

		Context("creating a user", Ordered, func() {
			var (
				login        string
				passwordHash []byte
			)

			expectCorrectUserData := func() {
				GinkgoHelper()

				users, err := query.Where(generated.User.Login.Eq(login)).Find(GinkgoT().Context())
				Expect(err).To(Succeed())
				Expect(users).To(HaveLen(1))
				Expect(users[0].ID).To(BeNumerically(">", 0))
				Expect(users[0].Login).To(Equal(login))
				Expect(users[0].PasswordHash).To(Equal(passwordHash))
			}

			BeforeAll(func() {
				login = exampleUserLogin
				passwordHash = generateRandomBytes()
			})

			Context("a new user", func() {
				It("should be created", func() {
					created, err := storage.CreateUser(
						GinkgoT().Context(),
						domain.UserID(login),
						domain.PasswordHash(passwordHash),
					)
					Expect(err).To(Succeed())
					Expect(created).To(BeTrue(), "user should be created")

					expectCorrectUserData()
				})
			})

			Context("another user with the same login", func() {
				It("should return ErrUserIDConflict", func() {
					created, err := storage.CreateUser(
						GinkgoT().Context(),
						domain.UserID(login),
						domain.PasswordHash(generateRandomBytes()),
					)
					Expect(err).To(MatchError(domain.ErrUserIDConflict))
					Expect(created).To(BeFalse(), "user should not be created")

					expectCorrectUserData()
				})
			})

			Context("requesting user's password hash", func() {
				It("should return the original hash", func() {
					hash, err := storage.GetPasswordHash(GinkgoT().Context(), domain.UserID(login))
					Expect(err).To(Succeed())
					Expect(hash.Bytes()).To(Equal(passwordHash))
				})
			})
		})

		Context("requesting user's password hash for non-existent user", func() {
			It("should return nil", func() {
				hash, err := storage.GetPasswordHash(GinkgoT().Context(), domain.UserID("non-existent-user"))
				Expect(err).To(Succeed())
				Expect(hash).To(BeNil())
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func ensureUserExists(storage *sqldatabase.Storage, login string) models.User {
	GinkgoHelper()

	created, err := storage.CreateUser(
		GinkgoT().Context(),
		domain.UserID(login),
		domain.PasswordHash(generateRandomBytes()),
	)
	Expect(created).To(BeTrue(), "user should be created")
	Expect(err).To(Succeed())

	user, err := sqldatabase.Query[models.User](storage).
		Where(generated.User.Login.Eq(login)).
		First(GinkgoT().Context())
	Expect(err).To(Succeed())

	return user
}

func generateRandomBytes() []byte {
	GinkgoHelper()

	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	Expect(err).To(Succeed())

	return buf
}
