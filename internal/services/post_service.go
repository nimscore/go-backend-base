package services

import (
	"context"

	"github.com/google/uuid"
	"go-backend-base/internal/orm"
)

// PostService defines the interface for post-related operations.
type PostService interface {
	CreatePost(ctx context.Context, communityID, authorID uuid.UUID, title, content string) (*orm.Post, error)
	GetPost(ctx context.Context, postID uuid.UUID) (*orm.Post, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, title, content *string) (*orm.Post, error)
	DeletePost(ctx context.Context, postID uuid.UUID) error
	PublishPost(ctx context.Context, postID uuid.UUID) error
	UnpublishPost(ctx context.Context, postID uuid.UUID) error
	ListUserPosts(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]orm.Post, string, error)
	LikePost(ctx context.Context, postID, userID uuid.UUID) error
	UnlikePost(ctx context.Context, postID, userID uuid.UUID) error
	CreateBookmark(ctx context.Context, postID, userID uuid.UUID) error
	DeleteBookmark(ctx context.Context, postID, userID uuid.UUID) error
	ListBookmarks(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]orm.Post, string, error)
}
