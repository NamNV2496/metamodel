package repository

//go:generate metamodel -source=$GOFILE -destination=../generated/ -tag=json
type Feature struct {
	FeatureName string `json:"feature_name,omitempty"`
	ScenarioID  int    `json:"scenario_id"`
	Status      string `bson:"status"`
	Description string `json:"description,omitempty" bson:"desc"`
	IgnoreMe    string // no tag, should be ignored
	SkippedTag  string `json:"-"` // skip tag
}
