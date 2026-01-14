package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

const (
	tokenIssuer           = "GopherMart"
	tokenValidationLeeway = 5 * time.Second
)

// TokenManager implements [handler.TokenService] interface.
type TokenManager struct {
	secretKeyProvider SecretKeyProvider
	signingMethod     jwt.SigningMethod
	tokenLifetime     time.Duration
}

// NewTokenManager creates an instance of [TokenManager].
func NewTokenManager(secretKeyProvider SecretKeyProvider, tokenLifetime time.Duration) *TokenManager {
	return &TokenManager{
		secretKeyProvider: secretKeyProvider,
		signingMethod:     jwt.SigningMethodHS256,
		tokenLifetime:     tokenLifetime,
	}
}

// IssueToken creates a JWT token for a given [domain.UserID].
// It is a responsibility of the caller to ensure that the provided user ID
// is valid, registered with the system and properly authenticated.
func (m *TokenManager) IssueToken(ctx context.Context, userID domain.UserID) (domain.Token, error) {
	token := jwt.NewWithClaims(m.signingMethod, m.createClaims(userID))

	secretKey, err := m.secretKeyProvider.GetSecretKey(ctx)
	if err != nil {
		return domain.TokenEmptyValue, fmt.Errorf("cannot get secret key: %w", err)
	}

	tokenValue, err := token.SignedString(secretKey)
	if err != nil {
		return domain.TokenEmptyValue, fmt.Errorf("cannot sign token: %w", err)
	}

	return domain.Token(tokenValue), nil
}

// ValidateToken performs a validation of a provided token value
// and extracts a [domain.UserID] from it.
// It returns an error if token signature is invalid or token has expired.
func (m *TokenManager) ValidateToken(ctx context.Context, tokenValue domain.Token) (domain.UserID, error) {
	var claims jwt.RegisteredClaims

	token, err := jwt.ParseWithClaims(tokenValue.String(), &claims, m.getKeyFunc(ctx), m.getParseOptions()...)
	if err != nil {
		return domain.UserIDEmptyValue, fmt.Errorf("cannot parse token: %w", err)
	}

	if !token.Valid {
		return domain.UserIDEmptyValue, ErrTokenNotValid
	}

	return m.extractUserID(claims)
}

func (m *TokenManager) createClaims(userID domain.UserID) *jwt.RegisteredClaims {
	claims := &jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   string(userID),
		Audience:  jwt.ClaimStrings{},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenLifetime)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        "",
	}

	return claims
}

func (m *TokenManager) getKeyFunc(ctx context.Context) jwt.Keyfunc {
	return func(_ *jwt.Token) (any, error) {
		return m.secretKeyProvider.GetSecretKey(ctx)
	}
}

func (m *TokenManager) getParseOptions() []jwt.ParserOption {
	return []jwt.ParserOption{
		jwt.WithStrictDecoding(),
		jwt.WithValidMethods([]string{m.signingMethod.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(tokenValidationLeeway),
		jwt.WithIssuer(tokenIssuer),
	}
}

func (m *TokenManager) extractUserID(claims jwt.RegisteredClaims) (domain.UserID, error) {
	userID := domain.UserID(claims.Subject)

	if userID == domain.UserIDEmptyValue {
		return domain.UserIDEmptyValue, ErrTokenInvalidUserID
	}

	return userID, nil
}
