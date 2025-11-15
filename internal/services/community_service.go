package services

import (
	"context"

	"github.com/google/uuid"
	"go-backend-base/internal/orm"
)

// CommunityService defines the interface for community-related operations.
type CommunityService interface {
	CreateCommunity(ctx context.Context, name, description string, creatorID uuid.UUID) (*orm.Community, error)
	GetCommunity(ctx context.Context, communityID uuid.UUID) (*orm.Community, error)
	UpdateCommunity(ctx context.Context, communityID uuid.UUID, name, description *string) (*orm.Community, error)
	DeleteCommunity(ctx context.Context, communityID uuid.UUID) error
	ListCommunities(ctx context.Context, cursor string, limit int) ([]orm.Community, string, error)
	ListUserCommunities(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]orm.Community, string, error)
	TransferOwnership(ctx context.Context, communityID, newOwnerID uuid.UUID) error
	JoinCommunity(ctx context.Context, communityID, userID uuid.UUID) error
	LeaveCommunity(ctx context.Context, communityID, userID uuid.UUID) error
	BanCommunity(ctx context.Context, communityID uuid.UUID, reason string) error
	UnbanCommunity(ctx context.Context, communityID uuid.UUID) error
}
