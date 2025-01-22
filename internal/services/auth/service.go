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

type AuthService struct {
	log             *slog.Logger
	storage         Storage
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewAuthService(
	log *slog.Logger,
	storage Storage,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:             log,
		storage:         storage,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (a *AuthService) Register(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

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

func (a *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := a.storage.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", sl.Err(err))
			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		log.Error("invalid credentials", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	accessToken, err := jwt.NewToken(user, a.accessTokenTTL, "access")
	refreshToken, err := jwt.NewToken(user, a.refreshTokenTTL, "refresh")
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (a *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (string, string, error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op))

	refreshPayload, err := jwt.ParseToken(refreshToken)
	if err != nil {
		log.Error("failed to parse token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if time.Now().Unix() > refreshPayload.Exp {
		log.Error("token expired", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.storage.GetUser(ctx, refreshPayload.Email)
	if err != nil {
		log.Error("user not found", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newAccessToken, err := jwt.NewToken(user, a.accessTokenTTL, "access")
	if err != nil {
		log.Error("can not gen new access token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newRefreshToken, err := jwt.NewToken(user, a.refreshTokenTTL, "refresh")
	if err != nil {
		log.Error("can not gen new refresh token", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return newAccessToken, newRefreshToken, nil
}
