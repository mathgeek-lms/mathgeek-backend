package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenService_GenerateAndValidateAccessToken(t *testing.T) {
	tokenService := NewTokenService("secret")

	accessToken, err := tokenService.GenerateAccessToken(1, "vasya@example.com", "STUDENT")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken.AccessToken)
	require.WithinDuration(t, time.Now().Add(24*time.Hour), accessToken.ExpiresAt, time.Second)

	claims, err := tokenService.ValidateAccessToken(accessToken.AccessToken)
	require.NoError(t, err)
	require.Equal(t, int64(1), claims.UserID)
	require.Equal(t, "STUDENT", claims.Role)
	require.Equal(t, "mathgeek-lms", claims.Issuer)
	require.Equal(t, "1", claims.Subject)
	require.Contains(t, claims.Audience, "mathgeek-lms-api")
	require.NotEmpty(t, claims.ID)
}

func TestTokenService_ValidateAccessToken_InvalidSecret(t *testing.T) {
	tokenService := NewTokenService("secret")
	otherTokenService := NewTokenService("other-secret")

	accessToken, err := tokenService.GenerateAccessToken(1, "vasya@example.com", "STUDENT")
	require.NoError(t, err)

	claims, err := otherTokenService.ValidateAccessToken(accessToken.AccessToken)

	require.Nil(t, claims)
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestTokenService_ValidateAccessToken_Expired(t *testing.T) {
	tokenService := NewTokenService("secret")
	token := newTestToken(t, "secret", jwt.RegisteredClaims{
		Issuer:    "mathgeek-lms",
		Subject:   "1",
		Audience:  jwt.ClaimStrings{"mathgeek-lms-api"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ID:        uuid.New().String(),
	})

	claims, err := tokenService.ValidateAccessToken(token)

	require.Nil(t, claims)
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestTokenService_ValidateAccessToken_NotReady(t *testing.T) {
	tokenService := NewTokenService("secret")
	token := newTestToken(t, "secret", jwt.RegisteredClaims{
		Issuer:    "mathgeek-lms",
		Subject:   "1",
		Audience:  jwt.ClaimStrings{"mathgeek-lms-api"},
		NotBefore: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.New().String(),
	})

	claims, err := tokenService.ValidateAccessToken(token)

	require.Nil(t, claims)
	require.ErrorIs(t, err, ErrTokenNotReady)
}

func TestTokenService_ValidateAccessToken_InvalidToken(t *testing.T) {
	tokenService := NewTokenService("secret")

	claims, err := tokenService.ValidateAccessToken("not-a-token")

	require.Nil(t, claims)
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func newTestToken(t *testing.T, secret string, registeredClaims jwt.RegisteredClaims) string {
	t.Helper()

	claims := Claims{
		UserID:           1,
		Role:             "STUDENT",
		RegisteredClaims: registeredClaims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	return signed
}

func TestTokenService_ValidateAccessToken_UnexpectedSigningMethod(t *testing.T) {
	tokenService := NewTokenService("secret")
	token := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		UserID: 1,
		Role:   "STUDENT",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "mathgeek-lms",
			Subject:   "1",
			Audience:  jwt.ClaimStrings{"mathgeek-lms-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	})
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	claims, err := tokenService.ValidateAccessToken(signed)

	require.Nil(t, claims)
	require.ErrorIs(t, err, ErrTokenInvalid)
}
