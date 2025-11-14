package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Bookmark struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	PostID    uuid.UUID
	Post      Post
	UserID    uuid.UUID
	User      User
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Bookmark) TableName() string {
	return "bookmark"
}

func (c *Bookmark) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectBookmarkByID(postID string, userID string) (*Bookmark, error) {
	var Bookmark Bookmark
	tx := c.database.
		Select(
			[]string{},
		).
		Where("post_id = ? AND user_id = ?", postID, userID).
		First(&Bookmark)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &Bookmark, nil
}

func (c *PostgresClient) SelectBookmarksWithPagination(limit int, cursor string) ([]*Bookmark, error) {
	var bookmarks []*Bookmark
	query := c.database.
		Select([]string{
			"post_id",
			"user_id",
		}).
		Preload("Post").
		Order("created_at DESC")

	if cursor != "" {
		var cursorBookmark Bookmark
		tx := c.database.
			Where("user_id = ?", cursor).
			First(&cursorBookmark)

		if tx.Error != nil {
			return nil, tx.Error
		}

		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorBookmark.CreatedAt,
			cursorBookmark.CreatedAt,
			cursorBookmark.ID,
		)
	}

	tx := query.Limit(limit).Find(&bookmarks)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return bookmarks, nil
}

func (c *PostgresClient) InsertBookmark(bookmark *Bookmark) error {
	tx := c.database.Create(bookmark)
	return tx.Error
}

func (c *PostgresClient) DeleteBookmark(bookmark *Bookmark) error {
	tx := c.database.Delete(bookmark)
	return tx.Error
}
