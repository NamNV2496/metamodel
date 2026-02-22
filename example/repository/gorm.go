package repository

import (
	"github.com/namnv2496/exmaple/entity"
)

//go:generate metamodel -source=$GOFILE -destination=../generated/ -tag=gorm -packageName=repository
type GormTest struct {
	entity.Entity  `gorm:"embedded"`
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

type EmbeddedEntity struct {
	entity.Entity `gorm:"embedded"`
	CategoryType  string `gorm:"column:category_type;not null;"`
	ParentId      uint32 `gorm:"column:parent_id;default:0;"`
	Value         uint32 `gorm:"column:value;not null;"`
}
