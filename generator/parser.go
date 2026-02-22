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

	// First pass: collect all struct types by name for embedded field resolution
	structTypes := make(map[string]*ast.StructType)
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
			structTypes[typeSpec.Name.Name] = structType
		}
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
				Fields:     parseFields(structType, tag, structTypes),
			}
			if len(meta.Fields) > 0 {
				structs = append(structs, meta)
			}
		}
	}

	return structs, packageName, nil
}

func parseFields(structType *ast.StructType, tag string, structTypes map[string]*ast.StructType) []FieldMeta {
	var fields []FieldMeta
	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}
		tagValue := strings.Trim(field.Tag.Value, "`")
		structTag := reflect.StructTag(tagValue)

		// Handle embedded structs (anonymous fields with gorm:"embedded")
		if len(field.Names) == 0 && tag == "gorm" {
			rawGorm := structTag.Get("gorm")
			if strings.Contains(rawGorm, "embedded") {
				// Resolve the embedded type name
				embeddedTypeName := ""
				switch t := field.Type.(type) {
				case *ast.Ident:
					embeddedTypeName = t.Name
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						embeddedTypeName = ident.Name
					}
				}
				if embeddedTypeName != "" {
					if embeddedStruct, ok := structTypes[embeddedTypeName]; ok {
						fields = append(fields, parseFields(embeddedStruct, tag, structTypes)...)
					}
				}
			}
			continue
		}

		tagName := parseTagName(structTag, tag)
		if tagName == "" || tagName == "-" || tagName == "->" {
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

func parseTagName(structTag reflect.StructTag, tagKey string) string {
	rawTag := structTag.Get(tagKey)
	if rawTag == "" {
		return ""
	}
	tag := strings.TrimSpace(rawTag)
	if tagKey == "gorm" {
		return parseGormTagName(structTag, tag)
	}
	// Split by comma to handle options like omitempty
	parts := strings.Split(rawTag, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return tag
}

func parseGormTagName(structTag reflect.StructTag, tag string) string {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
		if strings.HasPrefix(part, "many2many:") {
			return strings.TrimPrefix(part, "many2many:")
		}
		if strings.HasPrefix(part, "one2many:") {
			return strings.TrimPrefix(part, "one2many:")
		}
		// retry with json for id primary key
		rawTag := structTag.Get("json")
		if rawTag == "" {
			return ""
		}
		parts := strings.Split(rawTag, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}
