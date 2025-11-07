package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID `gorm:"primaryKey"`
	Slug              string
	Name              string
	Description       string
	Email             string
	Password          string
	Salt              string
	VerificationToken string
	ResetToken        string
	IsVerified        bool
	Communities       []Community `gorm:"foreignKey:OwnerID"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (c *User) TableName() string {
	return "user"
}

func (c *User) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectUserByID(ID string) (*User, error) {
	var user User
	tx := c.database.
		Select(
			[]string{
				"id",
				"slug",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"verification_token",
				"reset_token",
				"is_verified",
			},
		).
		Where("id = ?", ID).
		First(&user)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

func (c *PostgresClient) SelectUserBySlug(slug string) (*User, error) {
	var user User
	tx := c.database.
		Select(
			[]string{
				"id",
				"slug",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"verification_token",
				"reset_token",
				"is_verified",
			},
		).
		Where("slug = ?", slug).
		First(&user)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
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
				"verification_token",
				"reset_token",
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
				"slug",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"verification_token",
				"reset_token",
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

func (c *PostgresClient) SelectUserByVerificationToken(verificationToken string) (*User, error) {
	var user User
	tx := c.database.
		Select(
			[]string{
				"id",
				"slug",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"verification_token",
				"reset_token",
				"is_verified",
			},
		).
		Where("verification_token = ?", verificationToken).
		First(&user)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

func (c *PostgresClient) SelectUserByResetToken(resetToken string) (*User, error) {
	var user User
	tx := c.database.
		Select(
			[]string{
				"id",
				"slug",
				"name",
				"description",
				"email",
				"password",
				"salt",
				"verification_token",
				"reset_token",
				"is_verified",
			},
		).
		Where("reset_token = ?", resetToken).
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

func (c *PostgresClient) UpdateUser(user *User) error {
	tx := c.database.Model(user).Updates(user)
	return tx.Error
}
