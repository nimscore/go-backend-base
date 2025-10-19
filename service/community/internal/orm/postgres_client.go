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
	db, err := gorm.Open(
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

	raw, err := db.DB()
	if err != nil {
		return nil, err
	}

	raw.SetMaxOpenConns(1)
	raw.SetMaxIdleConns(1)
	raw.SetConnMaxIdleTime(5 * time.Second)

	return &PostgresClient{
		database: db,
	}, nil
}
