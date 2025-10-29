package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	ormpkg "github.com/stormhead-org/backend/internal/orm"
	"go.uber.org/zap"
)

var cleanupCommand = &cobra.Command{
	Use:   "cleanup",
	Short: "cleanup",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cleanupCommandImpl()
	},
}

func cleanupCommandImpl() error {
	var log *zap.Logger
	var err error
	if os.Getenv("DEBUG") == "1" {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}

	if err != nil {
		return err
	}

	if os.Getenv("DEBUG") == "1" {
		godotenv.Load()
	}

	log.Info("begin cleanup")

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
		return err
	}

	err = client.DeleteSessions()
	if err != nil {
		return err
	}

	log.Info("end cleanup")
	return nil
}

func init() {
	rootCommand.AddCommand(cleanupCommand)
}
