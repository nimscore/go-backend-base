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

func (User) TableName() string {
	return "users"
}

func (this *User) BeforeCreate(transaction *gorm.DB) error {
	this.ID = uuid.New()
	return nil
}

func (this *PostgresClient) SelectUserByUsername(username string) (*User, error) {
	var user User
	transaction := this.database.
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

func (this *PostgresClient) SelectUserByEmail(email string) (*User, error) {
	var user User
	transaction := this.database.
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

func (this *PostgresClient) InsertUser(user *User) error {
	transaction := this.database.Create(user)
	return transaction.Error
}
