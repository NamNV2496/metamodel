package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- toSnakeCase ----------------------------------------------------------------

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"User", "user"},
		{"UserName", "user_name"},
		{"GormTest", "gorm_test"},
		{"Id", "id"},
		{"UUID", "u_u_i_d"},
		{"userID", "user_i_d"},
		{"GetUserByID", "get_user_by_i_d"},
		{"HTTPServer", "h_t_t_p_server"},
		{"JSONResponse", "j_s_o_n_response"},
		{"ProductSKU", "product_s_k_u"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---- Generate (end-to-end) -------------------------------------------------------

const jsonFixture = `package models

type User struct {
	ID       int    ` + "`json:\"id\"`" + `
	Username string ` + "`json:\"username\"`" + `
	Email    string ` + "`json:\"email,omitempty\"`" + `
	Password string ` + "`json:\"-\"`" + `
}
`

const gormFixture = `package models

type Product struct {
	ID    uint    ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
	Name  string  ` + "`gorm:\"column:name\"`" + `
	Price float64 ` + "`gorm:\"column:price;not null\"`" + `
}
`

func TestGenerate_JSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "models.go")
	mustWriteFile(t, src, jsonFixture)

	cfg := Config{
		Source:      src,
		Destination: dir + "/",
		PackageName: "metamodel",
		Tag:         "json",
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := mustReadFile(t, filepath.Join(dir, "models_metamodel.go"))
	assertContains(t, content, `TableName: "users"`)
	assertContains(t, content, `FieldName: "id"`)
	assertContains(t, content, `FieldName: "username"`)
	assertContains(t, content, `FieldName: "email"`)
	// "-" tag must be excluded
	assertNotContains(t, content, `Password`)
}

func TestGenerate_GORM(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "models.go")
	mustWriteFile(t, src, gormFixture)

	cfg := Config{
		Source:      src,
		Destination: dir + "/",
		PackageName: "metamodel",
		Tag:         "gorm",
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := mustReadFile(t, filepath.Join(dir, "models_metamodel.go"))
	assertContains(t, content, `TableName: "products"`)
	assertContains(t, content, `FieldName: "name"`)
	assertContains(t, content, `FieldName: "price"`)
}

func TestGenerate_CustomTableName(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "models.go")
	mustWriteFile(t, src, jsonFixture)

	cfg := Config{
		Source:      src,
		Destination: dir + "/",
		PackageName: "metamodel",
		Tag:         "json",
		TableName:   "my_users",
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := mustReadFile(t, filepath.Join(dir, "models_metamodel.go"))
	assertContains(t, content, `TableName: "my_users"`)
}

func TestGenerate_DefaultDestination(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "models.go")
	mustWriteFile(t, src, jsonFixture)

	cfg := Config{Source: src, PackageName: "metamodel", Tag: "json"}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "models_metamodel.go")); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}

func TestGenerate_CreatesCommonAndOperatorFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "models.go")
	mustWriteFile(t, src, jsonFixture)

	cfg := Config{Source: src, Destination: dir + "/", PackageName: "metamodel", Tag: "json"}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, name := range []string{"common_metamodel.go", "gorm_operator_metamodel.go", "mongo_operator_metamodel.go"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to be created: %v", name, err)
		}
	}
}

func TestGenerate_NoStructsReturnsError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "empty.go")
	mustWriteFile(t, src, "package models\n")

	err := Generate(Config{Source: src, Tag: "json"})
	if err == nil {
		t.Fatal("expected error for file with no structs, got nil")
	}
}

func TestGenerate_ExplicitDestFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "models.go")
	dest := filepath.Join(dir, "out", "gen.go")
	mustWriteFile(t, src, jsonFixture)

	cfg := Config{Source: src, Destination: dest, PackageName: "metamodel", Tag: "json"}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Errorf("expected %s to exist: %v", dest, err)
	}
}

// ---- helpers --------------------------------------------------------------------

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func assertContains(t *testing.T, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("expected output to contain %q", substr)
	}
}

func assertNotContains(t *testing.T, content, substr string) {
	t.Helper()
	if strings.Contains(content, substr) {
		t.Errorf("expected output NOT to contain %q", substr)
	}
}
