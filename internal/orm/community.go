package orm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Community struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	OwnerID     uuid.UUID
	Owner       User
	Name        string
	Description string
	Slug        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *Community) TableName() string {
	return "communities"
}

func (c *Community) BeforeCreate(transaction *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}

func (c *PostgresClient) InsertCommunity(community *Community) error {
	transaction := c.database.Create(community)
	return transaction.Error
}

func (c *PostgresClient) SelectCommunityByID(id string) (*Community, error) {
	var community Community
	transaction := c.database.
		Select([]string{
			"id",
			"owner_id",
			"name",
			"description",
			"slug",
			"created_at",
			"updated_at",
		}).
		Where("id = ?", id).
		First(&community)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return &community, nil
}

func (c *PostgresClient) SelectCommunityBySlug(slug string) (*Community, error) {
	var community Community
	transaction := c.database.
		Select([]string{
			"id",
			"owner_id",
			"name",
			"description",
			"slug",
			"created_at",
			"updated_at",
		}).
		Where("slug = ?", slug).
		First(&community)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return &community, nil
}

func (c *PostgresClient) SelectCommunitiesWithPagination(limit int, cursor string) ([]*Community, error) {
	var communities []*Community

	query := c.database.
		Select([]string{
			"id",
			"owner_id",
			"name",
			"description",
			"slug",
			"created_at",
			"updated_at",
		}).
		Order("created_at DESC, id DESC")

	// Если передан cursor, используем его для пагинации
	if cursor != "" {
		var cursorCommunity Community
		if err := c.database.Where("id = ?", cursor).First(&cursorCommunity).Error; err == nil {
			// Продолжаем с элементов, которые созданы раньше или имеют меньший ID
			query = query.Where("(created_at < ?) OR (created_at = ? AND id < ?)",
				cursorCommunity.CreatedAt,
				cursorCommunity.CreatedAt,
				cursorCommunity.ID)
		}
	}

	transaction := query.Limit(limit).Find(&communities)

	if transaction.Error != nil {
		return nil, transaction.Error
	}

	return communities, nil
}
