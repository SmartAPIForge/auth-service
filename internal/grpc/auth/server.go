package authserver

import (
	"auth-service/internal/lib/jwt"
	authservice "auth-service/internal/services/auth"
	"auth-service/internal/storage"
	"context"
	"errors"
	authProto "github.com/SmartAPIForge/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	Login(
		ctx context.Context,
		email string,
		password string,
	) (accessToken string, refreshToken string, err error)
	Refresh(
		ctx context.Context,
		refreshToken string,
	) (string, string, error)
}

type AuthServer struct {
	authProto.UnimplementedAuthServer
	authService AuthService
}

func RegisterAuthServer(
	gRPCServer *grpc.Server,
	auth AuthService,
) {
	authProto.RegisterAuthServer(gRPCServer, &AuthServer{authService: auth})
}

func (s *AuthServer) Register(
	ctx context.Context,
	in *authProto.RegisterRequest,
) (*authProto.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.authService.Register(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &authProto.RegisterResponse{UserId: uid}, nil
}

func (s *AuthServer) Login(
	ctx context.Context,
	in *authProto.LoginRequest,
) (*authProto.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	accessToken, refreshToken, err := s.authService.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &authProto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) ValidateUser(
	_ context.Context,
	in *authProto.ValidateUserRequest,
) (*authProto.ValidateUserResponse, error) {
	response := &authProto.ValidateUserResponse{Valid: false}
	if in.AccessToken == "" {
		return response, status.Error(codes.InvalidArgument, "token missed")
	}

	payload, err := jwt.ParseToken(in.AccessToken)
	if err != nil {
		return response, status.Error(codes.Unauthenticated, "user unauthorized")
	}
	if !(payload.Role == in.RequiredRole) {
		return response, status.Error(codes.PermissionDenied, "role mismatch")
	}

	response.Valid = true
	return response, nil
}

func (s *AuthServer) Refresh(
	ctx context.Context,
	in *authProto.RefreshRequest,
) (*authProto.RefreshResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token missed")
	}

	accessToken, refreshToken, err := s.authService.Refresh(ctx, in.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to refresh")
	}

	return &authProto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
