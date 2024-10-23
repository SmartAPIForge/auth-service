package authservice

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/lib/jwt"
	"auth-service/internal/lib/sl"
	"auth-service/internal/storage"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Storage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	GetUser(ctx context.Context, email string) (models.User, error)
}

// impl

type AuthService struct {
	log      *slog.Logger
	storage  Storage
	tokenTTL time.Duration
}

func NewAuthService(
	log *slog.Logger,
	userRepo Storage,
	tokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:      log,
		storage:  userRepo,
		tokenTTL: tokenTTL,
	}
}

func (a *AuthService) Register(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.storage.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (a *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.storage.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in successfully")

	accessToken, err := jwt.NewAccessToken(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, nil
}
