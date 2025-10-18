package orm

import "github.com/google/uuid"

type Company struct {
	ID     uuid.UUID `gorm:"primaryKey,type:uuid;default:uuid_generate_v4()"`
	Name   string
	UserID uuid.UUID
	User   User
}

func (Company) TableName() string {
	return "companies"
}
