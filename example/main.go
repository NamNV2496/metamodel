package main

import (
	"fmt"

	metamodel_ "github.com/namnv2496/exmaple/generated"
	"gorm.io/gorm"
)

func main() {
	// Use the generated metamodel constants
	fmt.Println("Scenarios.TableName: ", metamodel_.Scenarios_.TableName)
	fmt.Println("Scenarios.Status: ", metamodel_.Scenarios_.Status)
	fmt.Println("Feature.ScenarioID: ", metamodel_.Feature_.ScenarioID)
	fmt.Println("AnotherModel.UserName: ", metamodel_.AnotherModel_.UserName)

	// build gorm query
	fmt.Println(metamodel_.GormTest_.FeatureName.Equal("1"))
	fmt.Println(metamodel_.GormTest_.FeatureName.Equal(2))
	fmt.Println(metamodel_.GormTest_.FeatureName.EqualString("5"))
	fmt.Println(metamodel_.GormTest_.FeatureName.EqualString(10))
	fmt.Println(metamodel_.GormTest_.FeatureName.WithDefaultOwner().Equal(1000))
	fmt.Println(metamodel_.GormTest_.FeatureName.WithOwner("features").Equal(1000))
	fmt.Println(metamodel_.GormTest_.FeatureName.WithOwnerString("features"))

	fmt.Println(metamodel_.Join("table", metamodel_.GormTest_.FeatureName.EqualString("1")))
	fmt.Println("select " + metamodel_.Columns(
		metamodel_.Scenarios_.Status.String(),
		metamodel_.Scenarios_.Description.String(),
	))

	var db *gorm.DB
	var results []map[string]interface{} // or []metamodel_.GormTest

	db.Table(metamodel_.GormTest_.TableName).
		Select(
			metamodel_.GormTest_.Id.String(),
			metamodel_.GormTest_.IsActive.String(),
			metamodel_.GormTest_.FeatureName.String(),
			metamodel_.GormTest_.PriceUnit.String(),
			metamodel_.GormTest_.Type.String(),
		).
		Where(metamodel_.GormTest_.IsActive.IsTrueString()).
		Where(metamodel_.GormTest_.FeatureName.EqualString("test")).
		Order(metamodel_.GormTest_.Id.AscString()).
		Find(&results)
	fmt.Println(db.Statement.SQL.String()) // or use db.Raw(query).Scan(&results)
	// or
	// query := metamodel_.NewQueryBuilder(metamodel_.GormTest_.TableName).
	// 	Select(
	// 		metamodel_.GormTest_.Id.String(),
	// 		metamodel_.GormTest_.IsActive.String(),
	// 		metamodel_.GormTest_.FeatureName.String(),
	// 		metamodel_.GormTest_.PriceUnit.String(),
	// 		metamodel_.GormTest_.Type.String(),
	// 	).
	// 	Where(
	// 		metamodel_.AndString(metamodel_.GormTest_.IsActive.IsTrueString()),
	// 		metamodel_.AndString(metamodel_.GormTest_.FeatureName.EqualString("test")),
	// 	).
	// 	OrderBy(metamodel_.GormTest_.Id.AscString()).
	// 	Build()
	// fmt.Println(query)
}
