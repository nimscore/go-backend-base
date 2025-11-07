package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentLike struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	CommentID uuid.UUID
	Comment   Comment
	UserID    uuid.UUID
	User      User
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *CommentLike) TableName() string {
	return "comment_like"
}

func (c *CommentLike) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectCommentLikeByID(commentID string, userID string) (*CommentLike, error) {
	var commentLike CommentLike
	tx := c.database.
		Select(
			[]string{},
		).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		First(&commentLike)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &commentLike, nil
}

func (c *PostgresClient) InsertCommentLike(commentLike *CommentLike) error {
	tx := c.database.Create(commentLike)
	return tx.Error
}

func (c *PostgresClient) DeleteCommentLike(commentLike *CommentLike) error {
	tx := c.database.Delete(commentLike)
	return tx.Error
}
