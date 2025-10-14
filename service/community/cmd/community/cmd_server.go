package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	grpcpkg "github.com/stormhead-org/service/iam/internal/grpc"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "server",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverCommandImpl()
	},
}

func serverCommandImpl() error {
	var logger interface{}
	if os.Getenv("DEBUG") == "1" {
		logger = zap.NewDevelopment
	} else {
		logger = zap.NewProduction
	}

	// Application
	application := fx.New(
		fx.NopLogger,
		fx.Provide(
			logger,

			// Application
			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) (*grpcpkg.GRPC, error) {
				grpcHost := os.Getenv("GQL_HOST")
				if grpcHost == "" {
					grpcHost = "127.0.0.1"
				}

				grpcPort := os.Getenv("GQL_PORT")
				if grpcPort == "" {
					grpcPort = "8080"
				}

				grpc := grpcpkg.NewGRPC(
					logger,
					grpcHost,
					grpcPort,
				)
				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return grpc.Start()
					},
					OnStop: func(ctx context.Context) error {
						return grpc.Stop()
					},
				})

				return grpc, nil
			},
		),
		fx.Invoke(
			func(*grpcpkg.GRPC) {
				// TODO:
			},
		),
	)
	application.Run()

	err := application.Err()
	if err != nil {
		os.Exit(1)
	}

	return nil
}

func init() {
	rootCommand.AddCommand(serverCommand)
}
