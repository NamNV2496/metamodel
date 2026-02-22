package repository

//go:generate metamodel -source=scenarios.go -destination=../generated/ -tag=bson -packageName=repository

type Scenarios struct {
	FeatureName string `json:"feature_name,omitempty"`
	ScenarioID  int    `json:"scenario_id"`
	Status      string `bson:"status"`
	Description string `json:"description,omitempty" bson:"desc"`
	IgnoreMe    string // no tag, should be ignored
	SkippedTag  string `json:"-"` // skip tag
}

type AnotherModel struct {
	UserID   string `json:"user_id"`
	UserName string `bson:"user_name"`
}
