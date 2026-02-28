package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// Config holds the configuration for code generation
type Config struct {
	Source      string
	Destination string
	PackageName string
	Tag         string
	TableName   string
}

// StructMeta holds metadata for a struct
type StructMeta struct {
	StructName string
	Fields     []FieldMeta
}

// FieldMeta holds metadata for a struct field
type FieldMeta struct {
	FieldName string
	TagName   string
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Generate generates metamodel code for the given configuration
func Generate(cfg Config) error {
	// Parse the source file
	structs, pkgName, err := parseFile(cfg.Source, cfg.Tag)
	if err != nil {
		return fmt.Errorf("failed to parse source file: %w", err)
	}
	if len(structs) == 0 {
		return fmt.Errorf("no structs found in %s", cfg.Source)
	}
	if cfg.PackageName != "" {
		pkgName = cfg.PackageName + "_"
	}
	destPath := cfg.Destination
	if destPath == "" {
		ext := filepath.Ext(cfg.Source)
		destPath = strings.TrimSuffix(cfg.Source, ext) + "_metamodel.go"
	} else if strings.HasSuffix(destPath, "/") || strings.HasSuffix(destPath, string(filepath.Separator)) {
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
		sourceBase := filepath.Base(cfg.Source)
		ext := filepath.Ext(sourceBase)
		filename := strings.TrimSuffix(sourceBase, ext) + "_metamodel.go"
		destPath = filepath.Join(destPath, filename)
	} else {
		destDir := filepath.Dir(destPath)
		if destDir != "." && destDir != "" {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
		}
	}

	// Prepare template data
	data := struct {
		PackageName string
		Structs     []StructMeta
	}{
		PackageName: pkgName,
		Structs:     structs,
	}
	// Execute template
	tableNameFn := func(structName string) string {
		if cfg.TableName != "" {
			return cfg.TableName
		}
		return toSnakeCase(structName) + "s"
	}
	tmpl, err := template.New("metamodel").Funcs(template.FuncMap{
		"tableName": tableNameFn,
	}).Parse(metamodelTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
	}
	// Write to destination file
	if err := os.WriteFile(destPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	destDir := filepath.Dir(destPath)
	if err := generateCommonFile(pkgName, destDir); err != nil {
		return fmt.Errorf("failed to write common file: %w", err)
	}
	// When tag is gorm, also generate the shared Field type file
	if err := generateOperatorFile(pkgName, destDir); err != nil {
		return fmt.Errorf("failed to write gorm field file: %w", err)
	}
	if err := generateMongoOperatorFile(pkgName, destDir); err != nil {
		return fmt.Errorf("failed to write mongo operator file: %w", err)
	}
	return nil
}

func generateCommonFile(pkgName, destDir string) error {
	tmpl, err := template.New("common").Parse(commonTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ PackageName string }{pkgName}); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
	}
	return os.WriteFile(filepath.Join(destDir, "common_metamodel.go"), formatted, 0644)
}

func generateMongoOperatorFile(pkgName, destDir string) error {
	filePath := filepath.Join(destDir, "mongo_operator_metamodel.go")
	tmpl, err := template.New("mongo_operator").Delims("[[", "]]").Parse(mongoFieldTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ PackageName string }{pkgName}); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
	}
	return os.WriteFile(filePath, formatted, 0644)
}

func generateOperatorFile(pkgName, destDir string) error {
	fieldFilePath := filepath.Join(destDir, "gorm_operator_metamodel.go")
	tmpl, err := template.New("operator").Parse(gormFieldTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ PackageName string }{pkgName}); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
	}
	return os.WriteFile(fieldFilePath, formatted, 0644)
}
