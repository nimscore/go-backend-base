package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var workerCommand = &cobra.Command{
	Use:   "worker",
	Short: "worker",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workerCommandImpl()
	},
}

func workerCommandImpl() error {
	var logger interface{}
	if os.Getenv("DEBUG") == "1" {
		logger = zap.NewDevelopment
	} else {
		logger = zap.NewProduction
	}

	// gqlHost := os.Getenv("GQL_HOST")
	// if gqlHost == "" {
	// 	gqlHost = "127.0.0.1"
	// }

	// gqlPort := os.Getenv("GQL_PORT")
	// if gqlPort == "" {
	// 	gqlPort = "8080"
	// }

	// Application
	application := fx.New(
		fx.NopLogger,
		fx.Provide(
			logger,

			// Application
			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) error {
				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})

				return nil
			},
		),
		fx.Invoke(
		// func(*apipkg.API) {
		// 	// TODO:
		// },
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
	rootCommand.AddCommand(workerCommand)
}
