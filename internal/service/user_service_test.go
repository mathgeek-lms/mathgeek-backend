package service

import (
	"context"
	"testing"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	repository_common "github.com/mathgeek-lms/mathgeek-backend/internal/repository/common"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository/mocks"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)

	phoneNumber := "+799912345678"
	normalizedPhoneNumber := "799912345678"
	request := model.CreateUserRequest{
		Name:        " vasya ",
		LastName:    " pupkin ",
		Email:       " VASYA@example.COM ",
		PhoneNumber: &phoneNumber,
		Password:    "12345678",
	}

	expectedRequest := model.CreateUserRequest{
		Name:        "vasya",
		LastName:    "pupkin",
		Email:       "vasya@example.com",
		PhoneNumber: &normalizedPhoneNumber,
		Password:    "12345678",
	}

	expectedResponse := model.CreateUserResponse{
		ID:          1,
		Name:        expectedRequest.Name,
		LastName:    expectedRequest.LastName,
		Email:       expectedRequest.Email,
		PhoneNumber: expectedRequest.PhoneNumber,
		Role:        "STUDENT",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repo.On("CreateUser", ctx, expectedRequest).Return(expectedResponse, nil)

	response, err := userService.CreateUser(ctx, request)

	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}

func TestUserService_LoginUser_UsesProvidedTokenGenerator(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)

	password := "12345678"
	passwordHash, err := repository_common.HashPassword(password)
	require.NoError(t, err)

	request := model.LoginUserRequest{
		Email:    " VASYA@example.COM ",
		Password: password,
	}

	user := model.User{
		ID:           1,
		Email:        "vasya@example.com",
		PasswordHash: passwordHash,
		Role:         "STUDENT",
	}

	expectedToken := AccessToken{
		AccessToken: "ready-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	repo.On("GetUserByEmail", ctx, "vasya@example.com").Return(user, nil)

	called := false
	tokenGenerator := tokenGeneratorFunc(func(userID int64, role string) (AccessToken, error) {
		called = true
		require.Equal(t, user.ID, userID)
		require.Equal(t, user.Role, role)
		return expectedToken, nil
	})

	accessToken, err := userService.LoginUser(ctx, request, tokenGenerator)

	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, expectedToken, accessToken)
}

type tokenGeneratorFunc func(userID int64, role string) (AccessToken, error)

func (f tokenGeneratorFunc) GenerateAccessToken(userID int64, role string) (AccessToken, error) {
	return f(userID, role)
}
