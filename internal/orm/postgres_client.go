package orm

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresClient struct {
	database *gorm.DB
}

func NewPostgresClient(host string, port string, user string, password string) (*PostgresClient, error) {
	database, err := gorm.Open(
		postgres.Open(
			fmt.Sprintf(
				"host=%s port=%s user=%s password=%s sslmode=disable",
				host,
				port,
				user,
				password,
			),
		),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}

	rawDatabase, err := database.DB()
	if err != nil {
		return nil, err
	}

	rawDatabase.SetMaxOpenConns(1)
	rawDatabase.SetMaxIdleConns(1)
	rawDatabase.SetConnMaxIdleTime(5 * time.Second)

	return &PostgresClient{
		database: database,
	}, nil
}
