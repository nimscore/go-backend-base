package grpc

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	eventpkg "github.com/stormhead-org/service/community/internal/event"
	jwtpkg "github.com/stormhead-org/service/community/internal/jwt"
	ormpkg "github.com/stormhead-org/service/community/internal/orm"
	protopb "github.com/stormhead-org/service/community/internal/proto"
	securitypkg "github.com/stormhead-org/service/community/internal/security"
)

var ErrUserExist = errors.New("user exist")

type AuthorizationServer struct {
	protopb.UnimplementedAuthorizationServiceServer
	jwt            *jwtpkg.JWT
	databaseClient *ormpkg.PostgresClient
	eventClient    *eventpkg.KafkaClient
}

func NewAuthorizationServer(jwt *jwtpkg.JWT, databaseClient *ormpkg.PostgresClient, eventClient *eventpkg.KafkaClient) *AuthorizationServer {
	return &AuthorizationServer{
		jwt:            jwt,
		databaseClient: databaseClient,
		eventClient:    eventClient,
	}
}

func (this *AuthorizationServer) Register(context context.Context, request *protopb.RegisterRequest) (*protopb.RegisterResponse, error) {
	_, err := this.databaseClient.SelectUserBySlug(
		request.Slug,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, ErrUserExist
	}

	_, err = this.databaseClient.SelectUserByEmail(
		request.Email,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, ErrUserExist
	}

	salt := securitypkg.GenerateSalt()

	hash, err := securitypkg.HashPassword(
		request.Password,
		salt,
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	err = this.eventClient.Write(
		context,
		eventpkg.AUTHORIZATION_REGISTER,
		eventpkg.AuthorizationRegisterMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *AuthorizationServer) Login(context context.Context, request *protopb.LoginRequest) (*protopb.LoginResponse, error) {
	user, err := this.databaseClient.SelectUserByEmail(
		request.Email,
	)
	if err != nil {
		return nil, err
	}

	if !user.IsVerified {
		return nil, errors.New("user not verified")
	}

	err = securitypkg.ComparePassword(
		user.Password,
		request.Password,
		user.Salt,
	)
	if err != nil {
		return nil, err
	}

	accessToken, err := this.jwt.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := this.jwt.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	err = this.eventClient.Write(
		context,
		eventpkg.AUTHORIZATION_LOGIN,
		eventpkg.AuthorizationLoginMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &protopb.LoginResponse{
		User: &protopb.User{
			Id:         user.ID.String(),
			Slug:       user.Slug,
			Email:      user.Email,
			IsVerified: user.IsVerified,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (this *AuthorizationServer) Logout(context context.Context, request *protopb.LogoutRequest) (*emptypb.Empty, error) {
	id, err := this.jwt.ParseRefreshToken(
		request.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	err = this.eventClient.Write(
		context,
		eventpkg.AUTHORIZATION_LOGOUT,
		eventpkg.AuthorizationLogoutMessage{
			ID: id,
		},
	)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (this *AuthorizationServer) RefreshToken(context context.Context, request *protopb.RefreshRequest) (*protopb.RefreshResponse, error) {
	id, err := this.jwt.ParseRefreshToken(
		request.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	accessToken, err := this.jwt.GenerateAccessToken(id)
	if err != nil {
		return nil, err
	}

	refreshToken, err := this.jwt.GenerateRefreshToken(id)
	if err != nil {
		return nil, err
	}

	err = this.eventClient.Write(
		context,
		eventpkg.AUTHORIZATION_REFRESH_TOKEN,
		eventpkg.AuthorizationRefreshTokenMessage{
			ID: id,
		},
	)
	if err != nil {
		return nil, err
	}

	return &protopb.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (this *AuthorizationServer) ValidateToken(context context.Context, request *protopb.ValidateRequest) (*protopb.ValidateResponse, error) {
	id, err := this.jwt.ParseAccessToken(
		request.AccessToken,
	)
	valid := err != nil

	err = this.eventClient.Write(
		context,
		eventpkg.AUTHORIZATION_VALIDATE_TOKEN,
		eventpkg.AuthorizationValidateTokenMessage{
			ID: id,
		},
	)
	if err != nil {
		return nil, err
	}

	return &protopb.ValidateResponse{
		Id:    id,
		Valid: valid,
	}, nil
}
