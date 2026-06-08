package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
	repository_common "github.com/mathgeek-lms/mathgeek-backend/internal/repository/common"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, request model.CreateUserRequest) (model.CreateUserResponse, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	if !emailRegex.MatchString(request.Email) {
		return model.CreateUserResponse{}, ErrInvalidEmail
	}

	_, err := validatePhone(request.PhoneNumber)
	if err != nil {
		return model.CreateUserResponse{}, ErrInvalidPhoneNumber
	}

	if len(request.Password) < 8 {
		return model.CreateUserResponse{}, ErrPasswordTooShort
	}

	user, err := s.repo.CreateUser(ctx, request)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return model.CreateUserResponse{}, ErrEmailAlreadyTaken
		}
		return model.CreateUserResponse{}, err
	}

	return user, nil
}

func (s *UserService) LoginUser(ctx context.Context, request model.LoginUserRequest) (AccessToken, error) {
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	if !emailRegex.MatchString(request.Email) {
		return AccessToken{}, ErrInvalidEmail
	}

	user, err := s.repo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AccessToken{}, ErrIncorrectPassword
		}
		return AccessToken{}, err
	}

	if !repository_common.ComparePasswordHash(request.Password, user.PasswordHash) {
		return AccessToken{}, ErrIncorrectPassword
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	tokenService := NewTokenService(jwtSecret)
	accessToken, err := tokenService.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return AccessToken{}, err
	}

	return accessToken, nil

}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (model.CreateUserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.CreateUserResponse{}, ErrUserNotFound
		}
	}

	return userToResponse(user), nil
}

func validatePhone(phone *string) (*string, error) {
	if phone == nil {
		return nil, nil
	}

	p := strings.TrimSpace(*phone)

	if strings.HasPrefix(p, "+") {
		p = p[1:]
	}

	if matched, _ := regexp.MatchString(`^\d+$`, p); !matched {
		return nil, fmt.Errorf("phone must contain only digits")
	}

	if len(p) != 12 {
		return nil, fmt.Errorf("phone must be 12 digits length")
	}

	*phone = p

	return phone, nil
}

func userToResponse(user model.User) model.CreateUserResponse {
	return model.CreateUserResponse{
		ID:          user.ID,
		Name:        user.Name,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

var (
	ErrPasswordTooShort   = errors.New("password too short")
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrEmailAlreadyTaken  = fmt.Errorf("email already in use")
	ErrInvalidPhoneNumber = fmt.Errorf("invalid phone number")
	ErrIncorrectPassword  = fmt.Errorf("incorrect password")
)
