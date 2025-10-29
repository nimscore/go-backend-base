package middleware

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	jwtpkg "github.com/stormhead-org/backend/internal/jwt"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
)

func NewAuthorizationMiddleware(logger *zap.Logger, jwt *jwtpkg.JWT, database *ormpkg.PostgresClient) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request interface{},
		information *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		bypassCheck := map[string]bool{
			"/proto.AuthorizationService/ValidateName":  true,
			"/proto.AuthorizationService/ValidateEmail": true,
			"/proto.AuthorizationService/Register":      true,
			"/proto.AuthorizationService/Login":         true,
		}
		if bypassCheck[information.FullMethod] {
			return handler(ctx, request)
		}

		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("missing metadata")
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		header, ok := meta["authorization"]
		if !ok {
			logger.Error("missing authorization header")
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token")
		}
		if !strings.HasPrefix(header[0], "Bearer ") {
			logger.Error("missing bearer")
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token")
		}

		token := strings.TrimPrefix(header[0], "Bearer ")

		id, err := jwt.ParseAccessToken(token)
		if err != nil {
			logger.Error("invalid access token", zap.Error(err))
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		session, err := database.SelectSessionByID(id)
		if err != nil {
			logger.Error("database error", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "internal error")
		}

		err = database.UpdateSession(session)
		if err != nil {
			logger.Error("database error", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "internal error")
		}

		ctx = SetSessionID(ctx, id)
		ctx = SetUserID(ctx, session.UserID.String())

		return handler(
			ctx,
			request,
		)
	}
}
