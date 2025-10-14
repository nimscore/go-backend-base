package orm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"primaryKey,type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func Debug() {
	database, err := NewDatabase("127.0.0.1", "5432", "postgres", "postgres")
	if err != nil {
		panic(err)
	}

	var user User
	database.db.First(&user)
	fmt.Println(user)
}
