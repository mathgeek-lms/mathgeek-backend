package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AccessToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type TokenService struct {
	accessSecret []byte
	accessTTL    time.Duration
}

func NewTokenService(accessSecret string) *TokenService {
	return &TokenService{
		accessSecret: []byte(accessSecret),
		accessTTL:    24 * time.Hour,
	}
}

func (s *TokenService) GenerateAccessToken(userID int64, email, role string) (AccessToken, error) {
	expiresAt := time.Now().Add(s.accessTTL)

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "mathgeek-lms",
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  jwt.ClaimStrings{"mathgeek-lms-api"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	if err != nil {
		return AccessToken{}, err
	}

	var accessToken AccessToken
	accessToken.AccessToken = signed
	accessToken.ExpiresAt = expiresAt
	return accessToken, nil
}

func (s *TokenService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method:%v", t.Header["alg"])
		}
		return s.accessSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrTokenNotReady
		default:
			return nil, ErrTokenInvalid
		}
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrTokenNotReady = errors.New("token not valid yet")
)
