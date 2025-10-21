package grpc

import (
	"context"
	"errors"
	"regexp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	eventpkg "github.com/stormhead-org/service/community/internal/event"
	jwtpkg "github.com/stormhead-org/service/community/internal/jwt"
	ormpkg "github.com/stormhead-org/service/community/internal/orm"
	protopkg "github.com/stormhead-org/service/community/internal/proto"
	securitypkg "github.com/stormhead-org/service/community/internal/security"
)

var ErrUserExist = errors.New("user exist")
var ErrInvalid = errors.New("invalid")

func ValidateUserSlug(slug string) error {
	if len(slug) < 5 {
		return ErrInvalid
	}

	return nil
}

func ValidateUserEmail(email string) error {
	regex, err := regexp.Compile(`[a-z\-\_\.]+@[a-z\-\_\.]+`)
	if err != nil {
		return ErrInvalid
	}

	if !regex.MatchString(email) {
		return ErrInvalid
	}

	return nil
}

type AuthorizationServer struct {
	protopkg.UnimplementedAuthorizationServiceServer
	logger         *zap.Logger
	jwt            *jwtpkg.JWT
	databaseClient *ormpkg.PostgresClient
	brokerClient   *eventpkg.KafkaClient
}

func NewAuthorizationServer(logger *zap.Logger, jwt *jwtpkg.JWT, databaseClient *ormpkg.PostgresClient, brokerClient *eventpkg.KafkaClient) *AuthorizationServer {
	return &AuthorizationServer{
		logger:         logger,
		jwt:            jwt,
		databaseClient: databaseClient,
		brokerClient:   brokerClient,
	}
}

func (this *AuthorizationServer) Register(context context.Context, request *protopkg.RegisterRequest) (*protopkg.RegisterResponse, error) {
	err := ValidateUserSlug(request.Slug)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "slug not match conditions")
	}

	err = ValidateUserEmail(request.Email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "email not match conditions")
	}

	_, err = this.databaseClient.SelectUserBySlug(
		request.Slug,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "slug already exist")
	}

	_, err = this.databaseClient.SelectUserByEmail(
		request.Email,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "email already exist")
	}

	salt := securitypkg.GenerateSalt()

	hash, err := securitypkg.HashPassword(
		request.Password,
		salt,
	)
	if err != nil {
		this.logger.Error("error hashing password", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	user := &ormpkg.User{
		Slug:       request.Slug,
		Email:      request.Email,
		Password:   hash,
		Salt:       salt,
		IsVerified: false,
	}

	err = this.databaseClient.InsertUser(user)
	if err != nil {
		this.logger.Error("error inserting user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	err = this.brokerClient.WriteMessage(
		context,
		eventpkg.AUTHORIZATION_REGISTER,
		eventpkg.AuthorizationRegisterMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		this.logger.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return nil, nil
}

func (this *AuthorizationServer) Login(context context.Context, request *protopkg.LoginRequest) (*protopkg.LoginResponse, error) {
	user, err := this.databaseClient.SelectUserByEmail(
		request.Email,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user not found")
	}
	if !user.IsVerified {
		return nil, status.Errorf(codes.InvalidArgument, "user not verified")
	}

	err = securitypkg.ComparePassword(
		user.Password,
		request.Password,
		user.Salt,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "password invalid")
	}

	accessToken, err := this.jwt.GenerateAccessToken(user.ID.String())
	if err != nil {
		this.logger.Error("error generating access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	refreshToken, err := this.jwt.GenerateRefreshToken(user.ID.String())
	if err != nil {
		this.logger.Error("error generating refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	err = this.brokerClient.WriteMessage(
		context,
		eventpkg.AUTHORIZATION_LOGIN,
		eventpkg.AuthorizationLoginMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		this.logger.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.LoginResponse{
		User: &protopkg.User{
			Id:          user.ID.String(),
			Name:        user.Name,
			Description: user.Description,
			Slug:        user.Slug,
			Email:       user.Email,
			IsVerified:  user.IsVerified,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (this *AuthorizationServer) Logout(context context.Context, request *protopkg.LogoutRequest) (*emptypb.Empty, error) {
	id, err := this.jwt.ParseRefreshToken(
		request.RefreshToken,
	)
	if err != nil {
		this.logger.Debug("refresh token error", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "refresh token invalid")
	}

	err = this.brokerClient.WriteMessage(
		context,
		eventpkg.AUTHORIZATION_LOGOUT,
		eventpkg.AuthorizationLogoutMessage{
			ID: id,
		},
	)
	if err != nil {
		this.logger.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (this *AuthorizationServer) RefreshToken(context context.Context, request *protopkg.RefreshRequest) (*protopkg.RefreshResponse, error) {
	id, err := this.jwt.ParseRefreshToken(
		request.RefreshToken,
	)
	if err != nil {
		this.logger.Debug("refresh token error", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "refresh token invalid")
	}

	accessToken, err := this.jwt.GenerateAccessToken(id)
	if err != nil {
		this.logger.Error("error generating access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	refreshToken, err := this.jwt.GenerateRefreshToken(id)
	if err != nil {
		this.logger.Error("error generating refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	err = this.brokerClient.WriteMessage(
		context,
		eventpkg.AUTHORIZATION_REFRESH_TOKEN,
		eventpkg.AuthorizationRefreshTokenMessage{
			ID: id,
		},
	)
	if err != nil {
		this.logger.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (this *AuthorizationServer) ValidateToken(context context.Context, request *protopkg.ValidateRequest) (*protopkg.ValidateResponse, error) {
	id, err := this.jwt.ParseAccessToken(
		request.AccessToken,
	)
	valid := err != nil
	if err != nil {
		this.logger.Debug("access token error", zap.Error(err))
	}

	err = this.brokerClient.WriteMessage(
		context,
		eventpkg.AUTHORIZATION_VALIDATE_TOKEN,
		eventpkg.AuthorizationValidateTokenMessage{
			ID: id,
		},
	)
	if err != nil {
		this.logger.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.ValidateResponse{
		Id:    id,
		Valid: valid,
	}, nil
}
