package middleware

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	jwtpkg "github.com/stormhead-org/service/community/internal/jwt"
)

func NewAuthorizationMiddleware(logger *zap.Logger, jwt *jwtpkg.JWT) grpc.UnaryServerInterceptor {
	return func(
		context context.Context,
		request interface{},
		information *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		bypassCheck := map[string]bool{
			"/proto.AuthorizationService/Register": true,
			"/proto.AuthorizationService/Login":    true,
		}
		if bypassCheck[information.FullMethod] {
			return handler(context, request)
		}

		meta, ok := metadata.FromIncomingContext(context)
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

		return handler(
			SetUserID(context, id),
			request,
		)
	}
}
