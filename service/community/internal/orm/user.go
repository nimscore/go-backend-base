package orm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	Slug       string
	Email      string
	Password   string
	Salt       string
	IsVerified bool
	Companies  []Company
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (User) TableName() string {
	return "users"
}

func (this *User) BeforeCreate(transaction *gorm.DB) error {
	this.ID = uuid.New()
	return nil
}

func (this *Database) SelectUserBySlug(slug string) (*User, error) {
	var user User
	transaction := this.db.
		Select(
			[]string{
				"id",
				"slug",
				"email",
				"password",
				"salt",
				"is_verified",
			},
		).
		Where("slug = ?", slug).
		First(&user)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return &user, nil
}

func (this *Database) SelectUserByEmail(slug string) (*User, error) {
	var user User
	transaction := this.db.
		Select(
			[]string{
				"id",
				"slug",
				"email",
				"password",
				"salt",
				"is_verified",
			},
		).
		Where("email = ?", slug).
		First(&user)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return &user, nil
}

func (this *Database) InsertUser(user *User) error {
	transaction := this.db.Create(&user)
	return transaction.Error
}

func Debug() {
	database, err := NewDatabase("127.0.0.1", "5432", "postgres", "postgres")
	if err != nil {
		panic(err)
	}

	var user User
	tx := database.db.Preload("Companies").First(&user)
	if tx.Error != nil {
		fmt.Println(tx.Error)
	}

	tx = database.db.Create(&User{
		ID:   uuid.New(),
		Slug: "userx",
	})

	if tx.Error != nil {
		fmt.Println(tx.Error)
	}
	// fmt.Println(user.Companies[0].Name)
}
