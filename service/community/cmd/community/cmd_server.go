package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	gqlpkg "github.com/stormhead-org/backend/service/community/internal/gql"
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
			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) (*gqlpkg.GQL, error) {
				gqlHost := os.Getenv("GQL_HOST")
				if gqlHost == "" {
					gqlHost = "127.0.0.1"
				}

				gqlPort := os.Getenv("GQL_PORT")
				if gqlPort == "" {
					gqlPort = "8080"
				}

				gql := gqlpkg.NewGQL(
					logger,
					gqlHost,
					gqlPort,
				)
				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return gql.Start()
					},
					OnStop: func(ctx context.Context) error {
						return gql.Stop()
					},
				})

				return gql, nil
			},
		),
		fx.Invoke(
			func(*gqlpkg.GQL) {
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
