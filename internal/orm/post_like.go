package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostLike struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	PostID    uuid.UUID
	Post      Post
	UserID    uuid.UUID
	User      User
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *PostLike) TableName() string {
	return "post_like"
}

func (c *PostLike) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectPostLike(postID string, userID string) (*PostLike, error) {
	var PostLike PostLike
	tx := c.database.
		Select(
			[]string{
				"id",
			},
		).
		Where("post_id = ? AND user_id = ?", postID, userID).
		First(&PostLike)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &PostLike, nil
}

func (c *PostgresClient) InsertPostLike(PostLike *PostLike) error {
	tx := c.database.Create(PostLike)
	return tx.Error
}

func (c *PostgresClient) DeletePostLike(PostLike *PostLike) error {
	tx := c.database.Delete(PostLike)
	return tx.Error
}
