package orm

import (
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"primaryKey,type:uuid;default:uuid_generate_v4()"`
	Name      string
	Companies []Company
	// CreatedAt time.Time `gorm:"not null"`
	// UpdatedAt time.Time `gorm:"not null"`
}

func (User) TableName() string {
	return "users"
}

type Company struct {
	ID     uuid.UUID `gorm:"primaryKey,type:uuid;default:uuid_generate_v4()"`
	Name   string
	UserID uuid.UUID
	User   User
}

func (Company) TableName() string {
	return "companies"
}

func Debug() {
	database, err := NewDatabase("127.0.0.1", "5432", "postgres", "postgres")
	if err != nil {
		panic(err)
	}

	var user User
	database.db.Preload("Companies").First(&user)
	fmt.Println(user.Companies[0].Name)
}
