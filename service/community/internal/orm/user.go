package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Name        string
	Description string
	Slug        string
	Email       string
	Password    string
	Salt        string
	IsVerified  bool
	Companies   []Company
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (User) TableName() string {
	return "users"
}

func (this *User) BeforeCreate(transaction *gorm.DB) error {
	this.ID = uuid.New()
	return nil
}

func (this *PostgresClient) SelectUserBySlug(slug string) (*User, error) {
	var user User
	transaction := this.database.
		Select(
			[]string{
				"id",
				"name",
				"description",
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

func (this *PostgresClient) SelectUserByEmail(slug string) (*User, error) {
	var user User
	transaction := this.database.
		Select(
			[]string{
				"id",
				"name",
				"description",
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

func (this *PostgresClient) InsertUser(user *User) error {
	transaction := this.database.Create(user)
	return transaction.Error
}
