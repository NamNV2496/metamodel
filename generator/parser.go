package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

func ParseFile(filename string, tag string) ([]StructMeta, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse file %s: %w", filename, err)
	}
	var structs []StructMeta
	packageName := node.Name.Name + "_"
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			structName := typeSpec.Name.Name
			meta := StructMeta{
				StructName: structName,
				Fields:     parseFields(structType, tag),
			}
			if len(meta.Fields) > 0 {
				structs = append(structs, meta)
			}
		}
	}

	return structs, packageName, nil
}

func parseFields(structType *ast.StructType, tag string) []FieldMeta {
	var fields []FieldMeta
	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}
		tagValue := strings.Trim(field.Tag.Value, "`")
		structTag := reflect.StructTag(tagValue)
		rawTag := structTag.Get(tag)
		if rawTag == "" {
			continue
		}
		tagName := parseTagName(rawTag, tag)
		if tagName == "" || tagName == "-" {
			continue
		}
		for _, ident := range field.Names {
			fields = append(fields, FieldMeta{
				FieldName: ident.Name,
				TagName:   tagName,
			})
		}
	}
	return fields
}

func parseTagName(rawTag string, tagKey string) string {
	tag := strings.TrimSpace(rawTag)
	if tagKey == "gorm" {
		return parseGormTagName(tag)
	}
	// Split by comma to handle options like omitempty
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return tag
}

func parseGormTagName(tag string) string {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}
