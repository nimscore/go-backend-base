package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Follower struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	UserID     uuid.UUID
	User       User
	FollowerID uuid.UUID
	Follower   User
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (c *Follower) TableName() string {
	return "follower"
}

func (c *Follower) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectFollowerByID(userID string, followerID string) (*Follower, error) {
	var Follower Follower
	tx := c.database.
		Select(
			[]string{},
		).
		Where("user_id = ? AND follower_id = ?", userID, followerID).
		First(&Follower)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &Follower, nil
}

func (c *PostgresClient) SelectFollowersWithPagination(userID string, followerID string, limit int, cursor string) ([]*Follower, error) {
	var followers []*Follower
	query := c.database.
		Select([]string{
			"user_id",
			"follower_id",
		}).
		Preload("User").
		Preload("Follower").
		Order("created_at DESC")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if followerID != "" {
		query = query.Where("follower_id = ?", followerID)
	}

	if cursor != "" {
		var cursorFollower Follower
		tx := c.database.
			Where("user_id = ?", cursor).
			First(&cursorFollower)

		if tx.Error != nil {
			return nil, tx.Error
		}

		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorFollower.CreatedAt,
			cursorFollower.CreatedAt,
			cursorFollower.ID,
		)
	}

	tx := query.Limit(limit).Find(&followers)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return followers, nil
}

func (c *PostgresClient) InsertFollower(follower *Follower) error {
	tx := c.database.Create(follower)
	return tx.Error
}

func (c *PostgresClient) DeleteFollower(follower *Follower) error {
	tx := c.database.Delete(follower)
	return tx.Error
}
