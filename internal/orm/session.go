package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID
	User      User
	UserAgent string
	IpAddress string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Session) TableName() string {
	return "sessions"
}

func (s *Session) BeforeCreate(transaction *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectSessionByID(id string) (*Session, error) {
	var session Session
	tx := c.database.
		Select([]string{
			"id",
			"user_id",
			"user_agent",
			"ip_address",
			"created_at",
			"updated_at",
		}).
		Where("id = ?", id).
		First(&session)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &session, nil
}

func (c *PostgresClient) InsertSession(session *Session) error {
	tx := c.database.Create(session)
	return tx.Error
}

func (c *PostgresClient) UpdateSession(session *Session) error {
	tx := c.database.Model(session).Updates(session)
	return tx.Error
}

func (c *PostgresClient) DeleteSession(session *Session) error {
	tx := c.database.Delete(session)
	return tx.Error
}
