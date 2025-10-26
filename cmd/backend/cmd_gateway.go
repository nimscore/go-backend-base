package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	gatewaypkg "github.com/stormhead-org/backend/internal/gateway"
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

	// Создание Gateway модуля
	gateway, err := gatewaypkg.NewGateway(
		logger,
		gatewayHost,
		gatewayPort,
		grpcEndpoint,
	)
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}

	// Запуск Gateway
	err = gateway.Start()
	if err != nil {
		return fmt.Errorf("failed to start gateway: %w", err)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gateway server...")

	if err := gateway.Stop(); err != nil {
		logger.Error("Gateway shutdown failed", zap.Error(err))
		return err
	}

	logger.Info("Gateway server stopped")
	return nil
}

func init() {
	rootCommand.AddCommand(gatewayCommand)
}
