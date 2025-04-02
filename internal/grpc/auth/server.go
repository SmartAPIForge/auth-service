package authserver

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/lib/jwt"
	authservice "auth-service/internal/services/auth"
	"auth-service/internal/storage"
	"context"
	"errors"
	authProto "github.com/SmartAPIForge/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

type UserService interface {
	GetUsers(
		ctx context.Context,
		roleID *int64,
		nameStartsWith *string,
	) ([]models.User, error)
	GetUserByToken(
		ctx context.Context,
		accessToken string,
	) (models.User, error)
	DeleteUser(
		ctx context.Context,
		username string,
	) error
}

type AuthServer struct {
	authProto.UnimplementedAuthServer
	authService AuthService
	userService UserService
}

func RegisterAuthServer(
	gRPCServer *grpc.Server,
	auth AuthService,
	user UserService,
) {
	authProto.RegisterAuthServer(gRPCServer, &AuthServer{
		authService: auth,
		userService: user,
	})
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
		return response, nil
	}

	payload, err := jwt.ParseToken(in.AccessToken)
	if err != nil {
		return response, nil
	}
	if !(payload.Role == in.RequiredRole) {
		return response, nil
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

func (s *AuthServer) GetUsers(
	ctx context.Context,
	in *authProto.GetUsersRequest,
) (*authProto.GetUsersResponse, error) {
	var roleID *int64
	var nameStartsWith *string

	if in.RoleId != 0 {
		val := in.RoleId
		roleID = &val
	}

	if in.NameStartsWith != "" {
		val := in.NameStartsWith
		nameStartsWith = &val
	}

	users, err := s.userService.GetUsers(ctx, roleID, nameStartsWith)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get users")
	}

	var protoUsers []*authProto.User
	for _, user := range users {
		protoUsers = append(protoUsers, &authProto.User{
			Username: user.Username,
			Email:    user.Email,
			RoleId:   user.Role,
		})
	}

	return &authProto.GetUsersResponse{
		Users: protoUsers,
	}, nil
}

func (s *AuthServer) GetUserByToken(
	ctx context.Context,
	in *authProto.GetUserByTokenRequest,
) (*authProto.User, error) {
	if in.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, err := s.userService.GetUserByToken(ctx, in.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &authProto.User{
		Username: user.Username,
		Email:    user.Email,
		RoleId:   user.Role,
	}, nil
}

func (s *AuthServer) DeleteUser(
	ctx context.Context,
	in *authProto.DeleteUserRequest,
) (*emptypb.Empty, error) {
	if in.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	err := s.userService.DeleteUser(ctx, in.Username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}
