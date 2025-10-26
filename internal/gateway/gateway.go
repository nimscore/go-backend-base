package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	protopkg "github.com/stormhead-org/backend/internal/proto"
)

type Gateway struct {
	logger       *zap.Logger
	host         string
	port         string
	grpcEndpoint string
	server       *http.Server
}

func NewGateway(
	logger *zap.Logger,
	host string,
	port string,
	grpcEndpoint string,
) (*Gateway, error) {
	return &Gateway{
		logger:       logger,
		host:         host,
		port:         port,
		grpcEndpoint: grpcEndpoint,
	}, nil
}

func (this *Gateway) Start() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Создание gRPC-gateway mux
	mux := runtime.NewServeMux(
		runtime.WithHealthzEndpoint(nil),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Регистрация всех gRPC сервисов
	if err := this.registerServices(ctx, mux, opts); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}

	// Создание HTTP сервера с маршрутами
	httpMux := this.setupRoutes(mux)

	gatewayAddr := fmt.Sprintf("%s:%s", this.host, this.port)
	this.logger.Info("Starting HTTP gateway server",
		zap.String("address", gatewayAddr),
		zap.String("grpc_endpoint", this.grpcEndpoint),
	)

	this.server = &http.Server{
		Addr:    gatewayAddr,
		Handler: httpMux,
	}

	go func() {
		this.logger.Info("Gateway server started")
		if err := this.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			this.logger.Error("Gateway server stopped", zap.Error(err))
		}
	}()

	return nil
}

func (this *Gateway) Stop() error {
	if this.server != nil {
		this.logger.Info("Stopping gateway server")
		return this.server.Close()
	}
	return nil
}

func (this *Gateway) registerServices(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
	// Authorization Service
	err := protopkg.RegisterAuthorizationServiceHandlerFromEndpoint(ctx, mux, this.grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("register authorization service: %w", err)
	}
	this.logger.Info("Registered AuthorizationService")

	// Community Service
	err = protopkg.RegisterCommunityServiceHandlerFromEndpoint(ctx, mux, this.grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("register community service: %w", err)
	}
	this.logger.Info("Registered CommunityService")

	// TODO: Register remaining services:
	// - UserService
	// - PostService
	// - CommentService
	// - FeedService
	// - RoleService
	// - PermissionService
	// - PlatformService
	// - ModerationService
	// - ReportService
	// - MediaService
	// - NotificationService
	// - SearchService
	// - BadgeService

	return nil
}

func (this *Gateway) setupRoutes(grpcMux *runtime.ServeMux) *http.ServeMux {
	mux := http.NewServeMux()

	// API endpoints - проксируем к gRPC
	mux.Handle("/", grpcMux)

	// Swagger UI - интерактивная документация на /docs
	mux.HandleFunc("/docs/", httpSwagger.Handler(
		httpSwagger.URL("/api/swagger.json"),
	))

	// Swagger JSON спецификация
	mux.HandleFunc("/api/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "api/swagger/api.swagger.json")
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	this.logger.Info("Gateway routes configured",
		zap.String("docs", "/docs/"),
		zap.String("swagger_spec", "/api/swagger.json"),
		zap.String("health", "/health"),
	)

	return mux
}

