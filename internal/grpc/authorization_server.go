package grpc

import (
	"context"
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

const SESSIONS_PER_PAGE = 10

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
	_, err := s.database.SelectUserByName(
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
	// Validate request
	err := ValidateUserName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name not match conditions")
	}

	err = ValidateUserEmail(request.Email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "email not match conditions")
	}

	// Validate name
	_, err = s.database.SelectUserByName(
		request.Name,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "name already exist")
	}

	// Validate email
	_, err = s.database.SelectUserByEmail(
		request.Email,
	)
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.InvalidArgument, "email already exist")
	}

	// Salt password
	salt := securitypkg.GenerateSalt()

	hash, err := securitypkg.HashPassword(
		request.Password,
		salt,
	)
	if err != nil {
		s.log.Error("can't hash password", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Create user
	user := &ormpkg.User{
		Name:              request.Name,
		Email:             request.Email,
		Password:          hash,
		Salt:              salt,
		VerificationToken: securitypkg.GenerateToken(),
		IsVerified:        false,
	}
	err = s.database.InsertUser(user)
	if err != nil {
		s.log.Error("can't insert user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Write message to broker
	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_REGISTER,
		eventpkg.AuthorizationRegisterMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		s.log.Error("can't write to broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return nil, nil
}

func (s *AuthorizationServer) Login(ctx context.Context, request *protopkg.LoginRequest) (*protopkg.LoginResponse, error) {
	// Get user from database
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
		s.log.Error("can't obtainin user agent or ip address")
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Check existing sessions
	sessions, err := s.database.SelectSessionsByUserID(user.ID.String(), "", 0)
	if err != nil {
		s.log.Error("can't obtainin sessions from database")
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	for _, session := range sessions {
		if session.IpAddress != ipAddress {
			continue
		}

		if session.UserAgent != userAgent {
			continue
		}

		s.log.Error("multiple login attempt from same client")
		return nil, status.Errorf(codes.Internal, "multiple login attempt from same client")
	}

	// Create session
	session := ormpkg.Session{
		UserID:    user.ID,
		UserAgent: userAgent,
		IpAddress: ipAddress,
	}
	err = s.database.InsertSession(&session)
	if err != nil {
		s.log.Error("can't insert session", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(session.ID.String())
	if err != nil {
		s.log.Error("can't generate access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(session.ID.String())
	if err != nil {
		s.log.Error("can't generate refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Write message to broker
	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_LOGIN,
		eventpkg.AuthorizationLoginMessage{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		s.log.Error("can't write to broker", zap.Error(err))
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
	// Get current session
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

	// Delete session from database
	err = s.database.DeleteSession(session)
	if err != nil {
		s.log.Error("can't delete session from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Write message to broker
	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_LOGOUT,
		eventpkg.AuthorizationLogoutMessage{
			ID: session.UserID.String(),
		},
	)
	if err != nil {
		s.log.Error("can't write to broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.LogoutResponse{}, nil
}

func (s *AuthorizationServer) RefreshToken(ctx context.Context, request *protopkg.RefreshTokenRequest) (*protopkg.RefreshTokenResponse, error) {
	// Get token
	id, err := s.jwt.ParseRefreshToken(
		request.RefreshToken,
	)
	if err != nil {
		s.log.Error("can't parse refresh token", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "refresh token invalid")
	}

	// Recreate tokens
	accessToken, err := s.jwt.GenerateAccessToken(id)
	if err != nil {
		s.log.Error("can't generate access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(id)
	if err != nil {
		s.log.Error("can't generate refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Write message to broker
	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_REFRESH_TOKEN,
		eventpkg.AuthorizationRefreshTokenMessage{
			ID: id,
		},
	)
	if err != nil {
		s.log.Error("can't write to broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthorizationServer) VerifyEmail(ctx context.Context, request *protopkg.VerifyEmailRequest) (*protopkg.VerifyEmailResponse, error) {
	// Find user
	user, err := s.database.SelectUserByVerificationToken(request.Token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user not exist")
	}

	// Update user
	user.IsVerified = true

	err = s.database.UpdateUser(user)
	if err != nil {
		s.log.Error("can't update user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.VerifyEmailResponse{}, nil
}

func (s *AuthorizationServer) RequestPasswordReset(ctx context.Context, request *protopkg.RequestPasswordResetRequest) (*protopkg.RequestPasswordResetResponse, error) {
	// Get user from database
	user, err := s.database.SelectUserByEmail(
		request.Email,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user not found")
	}
	if !user.IsVerified {
		return nil, status.Errorf(codes.InvalidArgument, "user not verified")
	}

	// Update user
	user.ResetToken = securitypkg.GenerateToken()

	err = s.database.UpdateUser(user)
	if err != nil {
		s.log.Error("can't update user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Write message to broker
	err = s.broker.WriteMessage(
		ctx,
		eventpkg.AUTHORIZATION_REQUEST_PASSWORD_RESET,
		eventpkg.AuthorizationRequestPasswordReset{
			ID: user.ID.String(),
		},
	)
	if err != nil {
		s.log.Error("can't write to broker", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.RequestPasswordResetResponse{}, nil
}

func (s *AuthorizationServer) ConfirmPasswordReset(ctx context.Context, request *protopkg.ConfirmResetPasswordRequest) (*protopkg.ConfirmResetPasswordResponse, error) {
	// Find user
	user, err := s.database.SelectUserByResetToken(request.Token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user not exist")
	}

	// Salt password
	salt := securitypkg.GenerateSalt()

	hash, err := securitypkg.HashPassword(
		request.Password,
		salt,
	)
	if err != nil {
		s.log.Error("can't hash password", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Update user in database
	user.Password = hash
	user.Salt = salt

	err = s.database.UpdateUser(user)
	if err != nil {
		s.log.Error("can't update user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.ConfirmResetPasswordResponse{}, nil
}

func (s *AuthorizationServer) ChangePassword(ctx context.Context, request *protopkg.ChangePasswordRequest) (*protopkg.ChangePasswordResponse, error) {
	// Get current user
	userID, err := middlewarepkg.GetUserID(ctx)
	if err != nil {
		s.log.Error("can't get user from middleware", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	user, err := s.database.SelectUserByID(userID)
	if err != nil {
		s.log.Error("can't get user from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Check password
	err = securitypkg.ComparePasswords(
		user.Password,
		request.OldPassword,
		user.Salt,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "password invalid")
	}

	// Salt password
	salt := securitypkg.GenerateSalt()

	hash, err := securitypkg.HashPassword(
		request.NewPassword,
		salt,
	)
	if err != nil {
		s.log.Error("can't hash password", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Update user in database
	user.Password = hash
	user.Salt = salt

	err = s.database.UpdateUser(user)
	if err != nil {
		s.log.Error("can't update user in database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.ChangePasswordResponse{}, nil
}

func (s *AuthorizationServer) GetCurrentSession(ctx context.Context, request *protopkg.GetCurrentSessionRequest) (*protopkg.GetCurrentSessionResponse, error) {
	// Get current session
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

func (s *AuthorizationServer) ListActiveSessions(ctx context.Context, request *protopkg.ListActiveSessionsRequest) (*protopkg.ListActiveSessionsResponse, error) {
	// Get current session
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

	// Get user sessions
	sessions, err := s.database.SelectSessionsByUserID(session.UserID.String(), request.Cursor, SESSIONS_PER_PAGE+1)
	if err != nil {
		s.log.Error("can't get sessions from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Build result
	hasMore := len(sessions) > SESSIONS_PER_PAGE

	var nextCursor string = ""
	if hasMore && len(sessions) > 0 {
		nextCursor = sessions[len(sessions)-1].ID.String()
	}

	if len(sessions) == SESSIONS_PER_PAGE+1 {
		sessions = sessions[:len(sessions)-1]
	}

	var result []*protopkg.Session
	for _, session := range sessions {
		result = append(
			result,
			&protopkg.Session{
				SessionId: session.ID.String(),
				UserAgent: session.UserAgent,
				IpAddress: session.IpAddress,
				CreatedAt: timestamppb.New(session.CreatedAt),
				UpdatedAt: timestamppb.New(session.UpdatedAt),
			},
		)
	}

	return &protopkg.ListActiveSessionsResponse{
		Sessions:   result,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

func (s *AuthorizationServer) RevokeSession(ctx context.Context, request *protopkg.RevokeSessionRequest) (*protopkg.RevokeSessionResponse, error) {
	// Get current session
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

	// Get requested session
	requestedSession, err := s.database.SelectSessionByID(request.SessionId)
	if err != nil {
		s.log.Error("can't get session from database", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "internal error")
	}

	if userSession.UserID != requestedSession.UserID {
		s.log.Error("can't validate session ownership", zap.Error(err))
		return nil, status.Errorf(codes.PermissionDenied, "internal error")
	}

	// Delete session from database
	err = s.database.DeleteSession(requestedSession)
	if err != nil {
		s.log.Error("can't delete session from database", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &protopkg.RevokeSessionResponse{}, nil
}
