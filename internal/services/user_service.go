package services

import (
	"context"

	"github.com/google/uuid"
	"go-backend-base/internal/orm"
)

// UserService defines the interface for user-related operations.
type UserService interface {
	Register(ctx context.Context, email, password string) (*orm.User, error)
	Login(ctx context.Context, email, password string) (*orm.User, *orm.Session, error)
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	GetCurrentSession(ctx context.Context, sessionID uuid.UUID) (*orm.Session, error)
	ListActiveSessions(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]orm.Session, string, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
}
