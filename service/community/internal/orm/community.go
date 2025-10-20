package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Community struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	OwnerID     uuid.UUID
	Owner       User
	Name        string
	Description string
	Slug        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Community) TableName() string {
	return "communities"
}

func (this *Community) BeforeCreate(transaction *gorm.DB) error {
	this.ID = uuid.New()
	return nil
}

func (this *PostgresClient) InsertCommunity(community *Community) error {
	transaction := this.database.Create(community)
	return transaction.Error
}
