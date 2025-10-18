package grpc

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	jwtpkg "github.com/stormhead-org/service/community/internal/jwt"
	ormpkg "github.com/stormhead-org/service/community/internal/orm"
	iampb "github.com/stormhead-org/service/community/internal/proto"
	securitypkg "github.com/stormhead-org/service/community/internal/security"
)

var ErrUserExist = errors.New("user exist")

type IAMServer struct {
	iampb.UnimplementedIAMServiceServer
	jwt      *jwtpkg.JWT
	database *ormpkg.Database
}

func NewIAMServer(jwt *jwtpkg.JWT, database *ormpkg.Database) *IAMServer {
	return &IAMServer{
		jwt:      jwt,
		database: database,
	}
}

func (this *IAMServer) Register(context context.Context, request *iampb.RegisterRequest) (*iampb.RegisterResponse, error) {
	_, err := this.database.SelectUserBySlug(
		request.Email,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, ErrUserExist
	}

	_, err = this.database.SelectUserByEmail(
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

	err = this.database.InsertUser(
		&ormpkg.User{
			Slug:       request.Slug,
			Email:      request.Email,
			Password:   hash,
			Salt:       salt,
			IsVerified: false,
		},
	)
	if err != nil {
		return nil, err
	}

	// TODO: kafka message

	return nil, nil
}

func (this *IAMServer) Login(context context.Context, request *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	user, err := this.database.SelectUserByEmail(
		request.Email,
	)
	if err != nil {
		return nil, err
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

	// TODO: kafka message

	return &iampb.LoginResponse{
		User: &iampb.User{
			Id:         user.ID.String(),
			Slug:       user.Slug,
			Email:      user.Email,
			IsVerified: user.IsVerified,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (this *IAMServer) Logout(context context.Context, request *iampb.LogoutRequest) (*emptypb.Empty, error) {
	// TODO: kafka message
	return &emptypb.Empty{}, nil
}

func (this *IAMServer) RefreshToken(context context.Context, request *iampb.RefreshRequest) (*iampb.RefreshResponse, error) {
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

	// TODO: kafka message
	return &iampb.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (this *IAMServer) ValidateToken(context context.Context, request *iampb.ValidateRequest) (*iampb.ValidateResponse, error) {
	id, err := this.jwt.ParseAccessToken(
		request.AccessToken,
	)
	if err != nil {
		return &iampb.ValidateResponse{
			Id:    id,
			Valid: false,
		}, nil
	}

	// TODO: kafka message
	return &iampb.ValidateResponse{
		Id:    id,
		Valid: true,
	}, nil
}
