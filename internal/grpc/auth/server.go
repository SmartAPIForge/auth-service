package authserver

import (
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
	Login(
		ctx context.Context,
		email string,
		password string,
	) (token string, err error)
	Register(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
}

type AuthServer struct {
	authProto.UnimplementedAuthServer
	authService AuthService
}

func RegisterAuthServer(gRPCServer *grpc.Server, auth AuthService) {
	authProto.RegisterAuthServer(gRPCServer, &AuthServer{authService: auth})
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

	token, err := s.authService.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &authProto.LoginResponse{Token: token}, nil
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
