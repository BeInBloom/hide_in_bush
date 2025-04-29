package userservice

import (
	"errors"
	"fmt"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type (
	repo interface {
		CreateUser(user models.User) (string, error)
		GetUserByLogin(login string) (models.User, error)
		GetUserByID(userID string) (models.User, error)
	}

	auth interface {
		GenerateToken(user string) (string, error)
	}
)

type UserService struct {
	repo repo
	auth auth
}

func New(
	repo repo,
	auth auth,
) *UserService {
	return &UserService{
		repo: repo,
		auth: auth,
	}
}

func (u *UserService) Register(
	credentials models.UserCredentials,
) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		Login:    credentials.Login,
		Password: string(hashedPassword),
	}

	userID, err := u.repo.CreateUser(user)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return "", errors.Join(ErrUserAlreadyExists)
		}

		return "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := u.auth.GenerateToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (u *UserService) ValidateCredentials(
	credentials models.UserCredentials,
) (ok bool, err error) {
	user, err := u.repo.GetUserByLogin(credentials.Login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return false, errors.Join(ErrUserNotFound)
		}

		return false, fmt.Errorf("failed to get user: %w", err)
	}

	ok, err = u.isCorrectPassword(user, credentials)
	if err != nil {
		return false, fmt.Errorf("failed to validate password: %w", err)
	}
	if !ok {
		return false, errors.Join(ErrInvalidCredentials)
	}

	return true, nil
}

func (u *UserService) UserBalance(
	userID string,
) (models.Balance, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return models.Balance{}, errors.Join(ErrUserNotFound)
		}

		return models.Balance{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user.Balance, nil
}

func (u *UserService) isCorrectPassword(
	user models.User,
	credentials models.UserCredentials,
) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, ErrInvalidCredentials
		}
		return false, fmt.Errorf("password comparison error: %w", err)
	}

	return true, nil
}
