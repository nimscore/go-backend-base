package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommunityUser struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	CommunityID uuid.UUID
	Community   Community
	UserID      uuid.UUID
	User        User
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *CommunityUser) TableName() string {
	return "community_user"
}

func (c *CommunityUser) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) SelectCommunityUser(communityID string, userID string) (*CommunityUser, error) {
	var communityUser CommunityUser
	tx := c.database.
		Select(
			[]string{
				"id",
			},
		).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		First(&communityUser)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &communityUser, nil
}

func (c *PostgresClient) InsertCommunityUser(communityUser *CommunityUser) error {
	tx := c.database.Create(communityUser)
	return tx.Error
}

func (c *PostgresClient) DeleteCommunityUser(communityUser *CommunityUser) error {
	tx := c.database.Delete(communityUser)
	return tx.Error
}
