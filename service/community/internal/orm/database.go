package orm

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	mu sync.Mutex
	db *gorm.DB
}

func NewDatabase(host string, port string, user string, password string) (*Database, error) {
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

	raw.SetMaxOpenConns(10)
	raw.SetMaxIdleConns(100)
	raw.SetConnMaxIdleTime(5 * time.Second)

	return &Database{
		db: db,
	}, nil
}
