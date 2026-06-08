package service

import (
	"context"
	"testing"
	"time"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
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
