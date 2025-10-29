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
	Email       string
	Password    string
	Salt        string
	IsVerified  bool
	Communities []Community `gorm:"foreignKey:OwnerID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *User) TableName() string {
	return "users"
}

func (c *User) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectUserByUsername(username string) (*User, error) {
	var user User
	transaction := c.database.
		Select(
			[]string{
				"id",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"is_verified",
			},
		).
		Where("name = ?", username).
		First(&user)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return &user, nil
}

func (c *PostgresClient) SelectUserByEmail(email string) (*User, error) {
	var user User
	transaction := c.database.
		Select(
			[]string{
				"id",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"is_verified",
			},
		).
		Where("email = ?", email).
		First(&user)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return &user, nil
}

func (c *PostgresClient) InsertUser(user *User) error {
	transaction := c.database.Create(user)
	return transaction.Error
}
