package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

//go:generate metamodel -source=$GOFILE -destination=../generated/ -tag=gorm -packageName=metamodel

type Entity struct {
	Id        uint      `gorm:"primaryKey" json:"id,omitempty"`
	Uuid      uuid.UUID `gorm:"column:uuid;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;" json:"updated_at,omitempty"`
}
