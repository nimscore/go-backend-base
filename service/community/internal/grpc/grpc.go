package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	iampb "github.com/stormhead-org/service/community/internal/proto"
)

type GRPC struct {
	logger *zap.Logger
	host   string
	port   string
	server *grpc.Server
}

func NewGRPC(logger *zap.Logger, host string, port string) *GRPC {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
		// middleware.GRPCAuthRateLimitMiddleware,
		// middleware.GRPCAuthMiddleware,
		),
	)

	// IAM API
	iamServer := NewIAMServer()
	iampb.RegisterIAMServiceServer(server, iamServer)

	// Health API
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	// Reflection API
	reflection.Register(server)

	return &GRPC{
		logger: logger,
		host:   host,
		port:   port,
		server: server,
	}
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
