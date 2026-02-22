package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

//go:generate metamodel -source=$GOFILE -destination=../generated/ -tag=gorm -packageName=repository
type GormTest struct {
	Entity         `gorm:"embedded"`
	FeatureName    string            `gorm:"column:feature_name;not null"`
	Type           int               `gorm:"column:type;default:1"`
	IsActive       bool              `gorm:"column:is_active"`
	GormElement    []GormElement     `gorm:"column:gorm_element"`
	PriceUnit      string            `gorm:"column:price_unit;type:varchar(250);default:'đ';"`
	EmbeddedEntity []*EmbeddedEntity `gorm:"many2many:embedded_entity;"`
	IgnoreMe       uint32            `gorm:"->"`
}

type GormElement struct {
	Name string `gorm:"column:name;not null"`
}

type Entity struct {
	Id        uint      `gorm:"primaryKey" json:"id,omitempty"`
	Uuid      uuid.UUID `gorm:"column:uuid;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;" json:"updated_at,omitempty"`
}

type EmbeddedEntity struct {
	Entity       `gorm:"embedded"`
	CategoryType string `gorm:"column:category_type;not null;"`
	ParentId     uint32 `gorm:"column:parent_id;default:0;"`
	Value        uint32 `gorm:"column:value;not null;"`
}
