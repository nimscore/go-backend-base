package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID              uuid.UUID `gorm:"primaryKey"`
	ParentCommentID *uuid.UUID
	ParentComment   *Comment
	PostID          uuid.UUID
	Post            Post
	AuthorID        uuid.UUID
	Author          User
	Content         string
	LikeCount       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (c *Comment) TableName() string {
	return "comment"
}

func (c *Comment) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectCommentByID(id string) (*Comment, error) {
	var comment Comment
	tx := c.database.
		Select([]string{
			"id",
			"parent_comment_id",
			"post_id",
			"author_id",
			"content",
			"created_at",
			"updated_at",
		}).
		Where("id = ?", id).
		Preload("Post").
		Preload("Author").
		First(&comment)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &comment, nil
}

func (c *PostgresClient) SelectCommentsWithPagination(post_id string, author_id string, limit int, cursor string) ([]*Comment, error) {
	var comments []*Comment
	query := c.database.
		Select([]string{
			"id",
			"parent_comment_id",
			"post_id",
			"author_id",
			"content",
			"created_at",
			"updated_at",
		}).
		Preload("Post").
		Preload("Author").
		Order("created_at DESC")

	if post_id != "" {
		query = query.Where("post_id = ?", post_id)
	}

	if author_id != "" {
		query = query.Where("author_id = ?", author_id)
	}

	if cursor != "" {
		var cursorComment Comment
		tx := c.database.
			Where("id = ?", cursor).
			First(&cursorComment)

		if tx.Error != nil {
			return nil, tx.Error
		}

		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorComment.CreatedAt,
			cursorComment.CreatedAt,
			cursorComment.ID,
		)
	}

	tx := query.Limit(limit).Find(&comments)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return comments, nil
}

func (c *PostgresClient) InsertComment(comment *Comment) error {
	transaction := c.database.Create(comment)
	return transaction.Error
}

func (c *PostgresClient) UpdateComment(comment *Comment) error {
	tx := c.database.Model(comment).Omit("Post").Omit("Author").Updates(comment)
	return tx.Error
}

func (c *PostgresClient) DeleteComment(comment *Comment) error {
	tx := c.database.Delete(comment)
	return tx.Error
}
