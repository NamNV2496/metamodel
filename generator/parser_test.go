package generator

import (
	"reflect"
	"testing"
)

// ---- parseTagName ---------------------------------------------------------------

func TestParseTagName_JSON(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`json:"username"`, "username"},
		{`json:"email,omitempty"`, "email"},
		{`json:"-"`, "-"},
		{`json:""`, ""},
	}
	for _, tt := range tests {
		tag := reflect.StructTag(tt.raw)
		got := parseTagName(tag, "json")
		if got != tt.want {
			t.Errorf("parseTagName(%q, json) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

// ---- parseGormTagName -----------------------------------------------------------

func TestParseGormTagName(t *testing.T) {
	tests := []struct {
		name    string
		rawGorm string
		rawJSON string
		want    string
	}{
		{
			name:    "explicit column tag",
			rawGorm: "column:user_name",
			want:    "user_name",
		},
		{
			// The json fallback fires on the first non-matching part ("not null"),
			// so "column:price" is never reached when there is no json tag.
			name:    "column not first part — returns empty without json fallback",
			rawGorm: "not null;column:price",
			want:    "",
		},
		{
			name:    "primaryKey falls back to json tag",
			rawGorm: "primaryKey",
			rawJSON: "id,omitempty",
			want:    "id",
		},
		{
			name:    "no column and no json yields empty",
			rawGorm: "not null",
			want:    "",
		},
		{
			name:    "many2many",
			rawGorm: "many2many:user_roles",
			want:    "user_roles",
		},
		{
			name:    "one2many",
			rawGorm: "one2many:order_items",
			want:    "order_items",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := `gorm:"` + tt.rawGorm + `"`
			if tt.rawJSON != "" {
				raw += ` json:"` + tt.rawJSON + `"`
			}
			tag := reflect.StructTag(raw)
			got := parseTagName(tag, "gorm")
			if got != tt.want {
				t.Errorf("parseTagName(gorm:%q, json:%q) = %q, want %q",
					tt.rawGorm, tt.rawJSON, got, tt.want)
			}
		})
	}
}

// ---- ParseFile ------------------------------------------------------------------

func TestParseFile_JSON(t *testing.T) {
	dir := t.TempDir()
	src := dir + "/models.go"
	mustWriteFile(t, src, `package models

type Order struct {
	ID     int    `+"`json:\"order_id\"`"+`
	Status string `+"`json:\"status\"`"+`
	Hidden string `+"`json:\"-\"`"+`
}
`)

	structs, pkg, err := parseFile(src, "json")
	if err != nil {
		t.Fatalf("parseFile() error = %v", err)
	}
	if pkg != "models_" {
		t.Errorf("pkg = %q, want %q", pkg, "models_")
	}
	if len(structs) != 1 {
		t.Fatalf("got %d structs, want 1", len(structs))
	}
	s := structs[0]
	if s.StructName != "Order" {
		t.Errorf("StructName = %q, want Order", s.StructName)
	}
	// "-" tagged field must be excluded
	want := []FieldMeta{
		{FieldName: "ID", TagName: "order_id"},
		{FieldName: "Status", TagName: "status"},
	}
	if !reflect.DeepEqual(s.Fields, want) {
		t.Errorf("Fields = %+v, want %+v", s.Fields, want)
	}
}

func TestParseFile_GORM(t *testing.T) {
	dir := t.TempDir()
	src := dir + "/models.go"
	mustWriteFile(t, src, `package models

type Item struct {
	ID    uint    `+"`gorm:\"primaryKey\" json:\"id\"`"+`
	Name  string  `+"`gorm:\"column:item_name\"`"+`
	Price float64 `+"`gorm:\"column:price;not null\"`"+`
}
`)

	structs, _, err := parseFile(src, "gorm")
	if err != nil {
		t.Fatalf("parseFile() error = %v", err)
	}
	if len(structs) != 1 {
		t.Fatalf("got %d structs, want 1", len(structs))
	}
	want := []FieldMeta{
		{FieldName: "ID", TagName: "id"},
		{FieldName: "Name", TagName: "item_name"},
		{FieldName: "Price", TagName: "price"},
	}
	if !reflect.DeepEqual(structs[0].Fields, want) {
		t.Errorf("Fields = %+v, want %+v", structs[0].Fields, want)
	}
}

func TestParseFile_MultipleStructs(t *testing.T) {
	dir := t.TempDir()
	src := dir + "/models.go"
	mustWriteFile(t, src, `package models

type Foo struct {
	A string `+"`json:\"a\"`"+`
}

type Bar struct {
	B string `+"`json:\"b\"`"+`
}
`)

	structs, _, err := parseFile(src, "json")
	if err != nil {
		t.Fatalf("parseFile() error = %v", err)
	}
	if len(structs) != 2 {
		t.Errorf("got %d structs, want 2", len(structs))
	}
}

func TestParseFile_SkipsStructsWithNoMatchingTag(t *testing.T) {
	dir := t.TempDir()
	src := dir + "/models.go"
	mustWriteFile(t, src, `package models

type NoTags struct {
	Name string
}

type WithTags struct {
	Name string `+"`json:\"name\"`"+`
}
`)

	structs, _, err := parseFile(src, "json")
	if err != nil {
		t.Fatalf("parseFile() error = %v", err)
	}
	if len(structs) != 1 {
		t.Fatalf("got %d structs, want 1", len(structs))
	}
	if structs[0].StructName != "WithTags" {
		t.Errorf("got struct %q, want WithTags", structs[0].StructName)
	}
}

func TestParseFile_InvalidPath(t *testing.T) {
	_, _, err := parseFile("/nonexistent/path/file.go", "json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseFile_InvalidGoSyntax(t *testing.T) {
	dir := t.TempDir()
	src := dir + "/bad.go"
	mustWriteFile(t, src, "package models\nthis is not valid go {{{{")

	_, _, err := parseFile(src, "json")
	if err == nil {
		t.Fatal("expected error for invalid Go syntax, got nil")
	}
}
