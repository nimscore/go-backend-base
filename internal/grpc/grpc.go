package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	eventpkg "github.com/stormhead-org/backend/internal/event"
	jwtpkg "github.com/stormhead-org/backend/internal/jwt"
	middlewarepkg "github.com/stormhead-org/backend/internal/middleware"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
	protopkg "github.com/stormhead-org/backend/internal/proto"
)

type GRPC struct {
	logger *zap.Logger
	host   string
	port   string
	server *grpc.Server
}

func NewGRPC(
	logger *zap.Logger,
	host string,
	port string,
	jwt *jwtpkg.JWT,
	database *ormpkg.PostgresClient,
	broker *eventpkg.KafkaClient,
) (*GRPC, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middlewarepkg.NewAuthorizationMiddleware(
				logger,
				jwt,
				database,
			),
		),
	)

	// Authorization API
	authorizationServer := NewAuthorizationServer(logger, jwt, database, broker)
	protopkg.RegisterAuthorizationServiceServer(grpcServer, authorizationServer)

	// Community API
	communityServer := NewCommunityServer(logger, database, broker)
	protopkg.RegisterCommunityServiceServer(grpcServer, communityServer)

	// Post API
	postServer := NewPostServer(logger, database, broker)
	protopkg.RegisterPostServiceServer(grpcServer, postServer)

	// Comment API
	commentServer := NewCommentServer(logger, database, broker)
	protopkg.RegisterCommentServiceServer(grpcServer, commentServer)

	// Health API
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Reflection API
	reflection.Register(grpcServer)

	return &GRPC{
		logger: logger,
		host:   host,
		port:   port,
		server: grpcServer,
	}, nil
}

func (this *GRPC) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", this.host, this.port))
	if err != nil {
		return err
	}

	go func() {
		this.logger.Info("GRPC server started")
		err := this.server.Serve(listener)
		if err != nil {
			this.logger.Error("GRPC server stopped", zap.Error(err))
		}
	}()

	return nil
}

func (this *GRPC) Stop() error {
	this.server.GracefulStop()
	return nil
}
