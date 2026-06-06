package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mathgeek-lms/mathgeek-backend/internal/model"
	"github.com/mathgeek-lms/mathgeek-backend/internal/repository"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, request model.CreateUserRequest) (model.User, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	if !emailRegex.MatchString(request.Email) {
		return model.User{}, ErrInvalidEmail
	}

	_, err := validatePhone(request.PhoneNumber)
	if err != nil {
		return model.User{}, ErrInvalidPhoneNumber
	}

	if len(request.Password) < 8 {
		return model.User{}, ErrPasswordTooShort
	}

	user, err := s.repo.CreateUser(ctx, request)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return model.User{}, ErrEmailAlreadyTaken
		}
		return model.User{}, err
	}

	return user, nil
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

var (
	ErrPasswordTooShort   = errors.New("password too short")
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrEmailAlreadyTaken  = fmt.Errorf("email already in use")
	ErrInvalidPhoneNumber = fmt.Errorf("invalid phone number")
)
