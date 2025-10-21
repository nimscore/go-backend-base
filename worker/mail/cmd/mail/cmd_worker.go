package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	clientpkg "github.com/stormhead-org/worker/mail/internal/client"
	workerpkg "github.com/stormhead-org/worker/mail/internal/worker"
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

	if os.Getenv("DEBUG") == "1" {
		godotenv.Load()
	}

	// Application
	application := fx.New(
		fx.NopLogger,
		fx.Provide(
			logger,

			// Kafka client
			func(logger *zap.Logger) (*clientpkg.KafkaClient, error) {
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
					kafkaTopic = "mail"
				}

				kafkaGroup := os.Getenv("KAFKA_GROUP")
				if kafkaGroup == "" {
					kafkaGroup = "mail"
				}

				kafkaClient := clientpkg.NewKafkaClient(
					kafkaHost,
					kafkaPort,
					kafkaTopic,
					kafkaGroup,
				)
				return kafkaClient, nil
			},

			// Mail client
			func(logger *zap.Logger) (*clientpkg.MailClient, error) {
				smtpHost := os.Getenv("SMTP_HOST")
				if smtpHost == "" {
					smtpHost = "127.0.0.1"
				}

				smtpPort := os.Getenv("SMTP_PORT")
				if smtpPort == "" {
					smtpPort = "587"
				}

				smtpUser := os.Getenv("SMTP_USER")
				if smtpUser == "" {
					smtpUser = "user"
				}

				smtpPassword := os.Getenv("SMTP_PASSWORD")
				if smtpPassword == "" {
					smtpPassword = "password"
				}

				mailClient := clientpkg.NewMailClient(
					smtpHost,
					smtpPort,
					smtpUser,
					smtpPassword,
				)
				return mailClient, nil
			},

			// Application
			func(
				lifecycle fx.Lifecycle,
				shutdowner fx.Shutdowner,
				logger *zap.Logger,
				kafkaClient *clientpkg.KafkaClient,
				mailClient *clientpkg.MailClient,
			) (*workerpkg.Worker, error) {
				worker := workerpkg.NewWorker(logger, kafkaClient, mailClient)

				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return worker.Start()
					},
					OnStop: func(ctx context.Context) error {
						return worker.Stop()
					},
				})

				return worker, nil
			},
		),
		fx.Invoke(
			func(*clientpkg.KafkaClient) {
				// TODO:
			},
			func(*clientpkg.MailClient) {
				// TODO:
			},
			func(*workerpkg.Worker) {
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
	rootCommand.AddCommand(workerCommand)
}
