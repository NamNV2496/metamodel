package main

import (
	"fmt"

	repository_ "github.com/namnv2496/exmaple/generated"
)

func main() {
	// Use the generated metamodel constants
	fmt.Println("Scenarios.TableName: ", repository_.Scenarios_.TableName)
	fmt.Println("Scenarios.Status: ", repository_.Scenarios_.Status)
	fmt.Println("Feature.ScenarioID: ", repository_.Feature_.ScenarioID)
	fmt.Println("AnotherModel.UserName: ", repository_.AnotherModel_.UserName)

	// build gorm query
	fmt.Println(repository_.GormTest_.FeatureName.Equal("1"))
	fmt.Println(repository_.GormTest_.FeatureName.Equal(2))
	fmt.Println(repository_.GormTest_.FeatureName.EqualString("5"))
	fmt.Println(repository_.GormTest_.FeatureName.EqualString(10))
	fmt.Println(repository_.GormTest_.FeatureName.IsFalse())
	fmt.Println("==========")
}
