package grpc

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	eventpkg "github.com/stormhead-org/backend/internal/event"
	jwtpkg "github.com/stormhead-org/backend/internal/jwt"
	middlewarepkg "github.com/stormhead-org/backend/internal/middleware"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
	protopkg "github.com/stormhead-org/backend/internal/proto"
	securitypkg "github.com/stormhead-org/backend/internal/security"
)

var ErrUserExist = errors.New("user exist")
var ErrInvalid = errors.New("invalid")

func ValidateUserName(name string) error {
	if len(name) < 5 {
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
	log      *zap.Logger
	jwt      *jwtpkg.JWT
	database *ormpkg.PostgresClient
	broker   *eventpkg.KafkaClient
}

func NewAuthorizationServer(log *zap.Logger, jwt *jwtpkg.JWT, database *ormpkg.PostgresClient, broker *eventpkg.KafkaClient) *AuthorizationServer {
	return &AuthorizationServer{
		log:      log,
		jwt:      jwt,
		database: database,
		broker:   broker,
	}
}

func (s *AuthorizationServer) ValidateName(ctx context.Context, request *protopkg.ValidateNameRequest) (*protopkg.ValidateNameResponse, error) {
	_, err := s.database.SelectUserByUsername(
		request.Name,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "name already exist")
	}

	return &protopkg.ValidateNameResponse{}, nil
}

func (s *AuthorizationServer) ValidateEmail(ctx context.Context, request *protopkg.ValidateEmailRequest) (*protopkg.ValidateEmailResponse, error) {
	_, err := s.database.SelectUserByEmail(
		request.Email,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "email already exist")
	}

	return &protopkg.ValidateEmailResponse{}, nil
}

func (s *AuthorizationServer) Register(ctx context.Context, request *protopkg.RegisterRequest) (*protopkg.RegisterResponse, error) {
	err := ValidateUserName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name not match conditions")
	}

	err = ValidateUserEmail(request.Email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "email not match conditions")
	}

	_, err = s.database.SelectUserByUsername(
		request.Name,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "name already exist")
	}

	_, err = s.database.SelectUserByEmail(
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
		s.log.Error("error hashing password", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	user := &ormpkg.User{
		Name:       request.Name,
		Email:      request.Email,
		Password:   hash,
		Salt:       salt,
		IsVerified: false,
	}

	err = s.database.InsertUser(user)
	if err != nil {
		s.log.Error("error inserting user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_REGISTER,
		eventpkg.AuthorizationRegisterMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		s.log.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return nil, nil
}

func (s *AuthorizationServer) Login(ctx context.Context, request *protopkg.LoginRequest) (*protopkg.LoginResponse, error) {
	user, err := s.database.SelectUserByEmail(
		request.Email,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user not found")
	}
	if !user.IsVerified {
		return nil, status.Errorf(codes.InvalidArgument, "user not verified")
	}

	err = securitypkg.ComparePasswords(
		user.Password,
		request.Password,
		user.Salt,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "password invalid")
	}

	// Obtain user agent and ip address
	userAgent := "unknown"
	m, ok := metadata.FromIncomingContext(ctx)
	if ok {
		userAgent = strings.Join(m["user-agent"], "")
	}

	ipAddress := "unknown"
	p, ok := peer.FromContext(ctx)
	if ok {
		parts := strings.Split(p.Addr.String(), ":")
		if len(parts) == 2 {
			ipAddress = parts[0]
		}
	}

	if userAgent == "unknown" || ipAddress == "unknown" {
		s.log.Error("error obtaining user agent or ip address")
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Create session
	session := ormpkg.Session{
		UserID:    user.ID,
		UserAgent: userAgent,
		IpAddress: ipAddress,
	}
	err = s.database.InsertSession(&session)
	if err != nil {
		s.log.Error("error inserting session", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(session.ID.String())
	if err != nil {
		s.log.Error("error generating access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(session.ID.String())
	if err != nil {
		s.log.Error("error generating refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Write event to broker
	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_LOGIN,
		eventpkg.AuthorizationLoginMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		s.log.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.LoginResponse{
		User: &protopkg.User{
			Id:          user.ID.String(),
			Name:        user.Name,
			Description: user.Description,
			Email:       user.Email,
			IsVerified:  user.IsVerified,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthorizationServer) Logout(ctx context.Context, request *protopkg.LogoutRequest) (*protopkg.LogoutResponse, error) {
	// id, err := this.jwt.ParseRefreshToken(
	// 	request.RefreshToken,
	// )
	// if err != nil {
	// 	this.logger.Debug("refresh token error", zap.Error(err))
	// 	return nil, status.Errorf(codes.InvalidArgument, "refresh token invalid")
	// }

	err := s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_LOGOUT,
		eventpkg.AuthorizationLogoutMessage{
			ID: "", // TODO: id
		},
	)
	if err != nil {
		s.log.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.LogoutResponse{}, nil
}

func (s *AuthorizationServer) RefreshToken(ctx context.Context, request *protopkg.RefreshTokenRequest) (*protopkg.RefreshTokenResponse, error) {
	id, err := s.jwt.ParseRefreshToken(
		request.RefreshToken,
	)
	if err != nil {
		s.log.Debug("refresh token error", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "refresh token invalid")
	}

	accessToken, err := s.jwt.GenerateAccessToken(id)
	if err != nil {
		s.log.Error("error generating access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(id)
	if err != nil {
		s.log.Error("error generating refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_REFRESH_TOKEN,
		eventpkg.AuthorizationRefreshTokenMessage{
			ID: id,
		},
	)
	if err != nil {
		s.log.Error("error writing broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthorizationServer) VerifyEmail(context.Context, *protopkg.VerifyEmailRequest) (*protopkg.VerifyEmailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyEmail not implemented")
}

func (s *AuthorizationServer) RequestPasswordReset(context.Context, *protopkg.RequestPasswordResetRequest) (*protopkg.RequestPasswordResetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestPasswordReset not implemented")
}

func (s *AuthorizationServer) ConfirmPasswordReset(context.Context, *protopkg.ConfirmResetPasswordRequest) (*protopkg.ConfirmResetPasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmPasswordReset not implemented")
}

func (s *AuthorizationServer) ChangePassword(context.Context, *protopkg.ChangePasswordRequest) (*protopkg.ChangePasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}

func (s *AuthorizationServer) GetCurrentSession(ctx context.Context, request *protopkg.GetCurrentSessionRequest) (*protopkg.GetCurrentSessionResponse, error) {
	sessionID, err := middlewarepkg.GetSessionID(ctx)
	if err != nil {
		s.log.Error("can't get session from middleware", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	session, err := s.database.SelectSessionByID(sessionID)
	if err != nil {
		s.log.Error("can't get session from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.GetCurrentSessionResponse{
		Session: &protopkg.Session{
			SessionId: session.ID.String(),
			UserAgent: session.UserAgent,
			IpAddress: session.IpAddress,
			CreatedAt: timestamppb.New(session.CreatedAt),
			UpdatedAt: timestamppb.New(session.UpdatedAt),
		},
	}, nil
}

func (s *AuthorizationServer) ListActiveSessions(context.Context, *protopkg.ListActiveSessionsRequest) (*protopkg.ListActiveSessionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListActiveSessions not implemented")
}

func (s *AuthorizationServer) RevokeSession(ctx context.Context, request *protopkg.RevokeSessionRequest) (*protopkg.RevokeSessionResponse, error) {
	sessionID, err := middlewarepkg.GetSessionID(ctx)
	if err != nil {
		s.log.Error("can't get session from middleware", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	userSession, err := s.database.SelectSessionByID(sessionID)
	if err != nil {
		s.log.Error("can't get session from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	requestedSession, err := s.database.SelectSessionByID(request.SessionId)
	if err != nil {
		s.log.Error("can't get session from database", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "internal error")
	}

	if userSession.UserID != requestedSession.UserID {
		s.log.Error("can't validate session ownership", zap.Error(err))
		return nil, status.Errorf(codes.PermissionDenied, "internal error")
	}

	err = s.database.DeleteSession(requestedSession)
	if err != nil {
		s.log.Error("can't delete session from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.RevokeSessionResponse{}, nil
}
