package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	protopkg "github.com/stormhead-org/service/community/internal/proto"
)

var gatewayCommand = &cobra.Command{
	Use:   "gateway",
	Short: "Start HTTP/REST gateway server",
	Long:  "Start HTTP/REST gateway server with Swagger UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return gatewayCommandImpl()
	},
}

func gatewayCommandImpl() error {
	var logger *zap.Logger
	var err error
	
	if os.Getenv("DEBUG") == "1" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// HTTP gateway настройки
	gatewayHost := os.Getenv("GATEWAY_HOST")
	if gatewayHost == "" {
		gatewayHost = "0.0.0.0"
	}

	gatewayPort := os.Getenv("GATEWAY_PORT")
	if gatewayPort == "" {
		gatewayPort = "8090"
	}

	// gRPC сервер настройки (к которому будет подключаться gateway)
	grpcHost := os.Getenv("GRPC_HOST")
	if grpcHost == "" {
		grpcHost = "127.0.0.1"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "8080"
	}

	grpcEndpoint := fmt.Sprintf("%s:%s", grpcHost, grpcPort)

	// Создание gRPC-gateway mux
	mux := runtime.NewServeMux(
		runtime.WithHealthzEndpoint(nil),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Регистрация Authorization Service
	err = protopkg.RegisterAuthorizationServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		logger.Fatal("failed to register authorization service handler", zap.Error(err))
	}

	// Регистрация Community Service
	err = protopkg.RegisterCommunityServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		logger.Fatal("failed to register community service handler", zap.Error(err))
	}

	// Создание HTTP сервера с Swagger UI
	httpMux := http.NewServeMux()
	
	// API endpoints
	httpMux.Handle("/", mux)
	
	// Swagger UI - интерактивная документация
	httpMux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/api.swagger.json"),
	))
	
	// Swagger JSON спецификация
	httpMux.HandleFunc("/swagger/api.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "api/swagger/api.swagger.json")
	})

	// Health check
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Корневой маршрут - редирект на Swagger UI
	httpMux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	gatewayAddr := fmt.Sprintf("%s:%s", gatewayHost, gatewayPort)
	logger.Info("Starting HTTP gateway server",
		zap.String("address", gatewayAddr),
		zap.String("grpc_endpoint", grpcEndpoint),
	)

	server := &http.Server{
		Addr:    gatewayAddr,
		Handler: httpMux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("gateway server failed", zap.Error(err))
	}

	return nil
}

func init() {
	rootCommand.AddCommand(gatewayCommand)
}

