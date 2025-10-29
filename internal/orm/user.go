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

func (c *PostgresClient) SelectUserByName(name string) (*User, error) {
	var user User
	tx := c.database.
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
		Where("name = ?", name).
		First(&user)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

func (c *PostgresClient) SelectUserByEmail(email string) (*User, error) {
	var user User
	tx := c.database.
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

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

func (c *PostgresClient) InsertUser(user *User) error {
	tx := c.database.Create(user)
	return tx.Error
}
