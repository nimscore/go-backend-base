package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	eventpkg "github.com/stormhead-org/backend/internal/event"
	grpcpkg "github.com/stormhead-org/backend/internal/grpc"
	jwtpkg "github.com/stormhead-org/backend/internal/jwt"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
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

	if os.Getenv("DEBUG") == "1" {
		godotenv.Load()
	}

	// Application
	application := fx.New(
		// fx.NopLogger,
		fx.Provide(
			logger,

			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) (*jwtpkg.JWT, error) {
				jwtSecret := os.Getenv("JWT_SECRET")
				if jwtSecret == "" {
					jwtSecret = "123456"
				}

				return jwtpkg.NewJWT(jwtSecret), nil
			},

			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) (*ormpkg.PostgresClient, error) {
				postgresHost := os.Getenv("POSTGRES_HOST")
				if postgresHost == "" {
					postgresHost = "127.0.0.1"
				}

				postgresPort := os.Getenv("POSTGRES_PORT")
				if postgresPort == "" {
					postgresPort = "5432"
				}

				postgresUser := os.Getenv("POSTGRES_USER")
				if postgresUser == "" {
					postgresUser = "postgres"
				}

				postgresPassword := os.Getenv("POSTGRES_PASSWORD")
				if postgresPassword == "" {
					postgresPassword = "postgres"
				}

				client, err := ormpkg.NewPostgresClient(
					postgresHost,
					postgresPort,
					postgresUser,
					postgresPassword,
				)
				if err != nil {
					return nil, err
				}

				return client, err
			},

			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) (*eventpkg.KafkaClient, error) {
				kafkaHost := os.Getenv("KAFKA_HOST")
				if kafkaHost == "" {
					kafkaHost = "127.0.0.1"
				}

				kafkaPort := os.Getenv("KAFKA_PORT")
				if kafkaPort == "" {
					kafkaPort = "9092"
				}

				kafkaTopic := os.Getenv("KAFKA_TOPIC")
				if kafkaTopic == "" {
					kafkaTopic = "common"
				}

				kafkaGroup := os.Getenv("KAFKA_GROUP")
				if kafkaGroup == "" {
					kafkaGroup = "common"
				}

				client, err := eventpkg.NewKafkaClient(
					kafkaHost,
					kafkaPort,
					kafkaTopic,
					kafkaGroup,
				)
				return client, err
			},

			// Application
			func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger, jwt *jwtpkg.JWT, databaseClient *ormpkg.PostgresClient, brokerClient *eventpkg.KafkaClient) (*grpcpkg.GRPC, error) {
				grpcHost := os.Getenv("GRPC_HOST")
				if grpcHost == "" {
					grpcHost = "127.0.0.1"
				}

				grpcPort := os.Getenv("GRPC_PORT")
				if grpcPort == "" {
					grpcPort = "8080"
				}

				grpc, err := grpcpkg.NewGRPC(
					logger,
					grpcHost,
					grpcPort,
					jwt,
					databaseClient,
					brokerClient,
				)
				if err != nil {
					return nil, err
				}

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
			func(*jwtpkg.JWT) {
				// TODO:
			},
			func(*ormpkg.PostgresClient) {
				// TODO:
			},
			func(*eventpkg.KafkaClient) {
				// TODO:
			},
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
