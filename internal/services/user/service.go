package user

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/lib/jwt"
	"auth-service/internal/lib/sl"
	"auth-service/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Storage interface {
	GetUser(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, userID int64) (models.User, error)
	GetUsers(ctx context.Context, roleID *int64, nameStartsWith *string) ([]models.User, error)
	DeleteUserByUsername(ctx context.Context, username string) error
}

type UserService struct {
	log     *slog.Logger
	storage Storage
}

func NewUserService(
	log *slog.Logger,
	storage Storage,
) *UserService {
	return &UserService{
		log:     log,
		storage: storage,
	}
}

func (s *UserService) GetUsers(
	ctx context.Context,
	roleID *int64,
	nameStartsWith *string,
) ([]models.User, error) {
	const op = "user.GetUsers"

	log := s.log.With(
		slog.String("op", op),
	)

	users, err := s.storage.GetUsers(ctx, roleID, nameStartsWith)
	if err != nil {
		log.Error("failed to get users", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (s *UserService) GetUserByToken(
	ctx context.Context,
	accessToken string,
) (models.User, error) {
	const op = "user.GetUserByToken"

	log := s.log.With(
		slog.String("op", op),
	)

	payload, err := jwt.ParseToken(accessToken)
	if err != nil {
		log.Error("failed to parse token", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.storage.GetUserByID(ctx, payload.Uid)
	if err != nil {
		log.Error("failed to get user", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	username string,
) error {
	const op = "user.DeleteUser"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	err := s.storage.DeleteUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		log.Error("failed to delete user", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
