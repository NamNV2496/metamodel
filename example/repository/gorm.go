package repository

//go:generate metamodel -source=$GOFILE -destination=../generated/ -tag=gorm
type GormTest struct {
	FeatureName string        `gorm:"column:feature_name;not null"`
	Type        int           `gorm:"column:type;default:1"`
	IsActive    bool          `gorm:"column:is_active"`
	GormElement []GormElement `gorm:"column:gorm_element"`
}

type GormElement struct {
	Name string `gorm:"column:name;not null"`
}
