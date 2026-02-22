package generator

import (
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
		return nil, "", err
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
				Name:   structName,
				Fields: parseFields(structType, tag),
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
		// Parse the tag
		tagValue := strings.Trim(field.Tag.Value, "`")
		structTag := reflect.StructTag(tagValue)
		// Extract tag based on preference
		tagName := ""
		if tag == "json" {
			if jsonTag := structTag.Get("json"); jsonTag != "" {
				tagName = parseTagName(jsonTag)
			} else if bsonTag := structTag.Get("bson"); bsonTag != "" {
				tagName = parseTagName(bsonTag)
			}
		} else if tag == "bson" {
			if bsonTag := structTag.Get("bson"); bsonTag != "" {
				tagName = parseTagName(bsonTag)
			} else if jsonTag := structTag.Get("json"); jsonTag != "" {
				tagName = parseTagName(jsonTag)
			}
		}
		if tagName == "" || tagName == "-" {
			continue
		}
		for _, ident := range field.Names {
			fields = append(fields, FieldMeta{
				Name:    ident.Name,
				TagName: tagName,
			})
		}
	}
	return fields
}

func parseTagName(tag string) string {
	// Remove any spaces (some people write "json: field_name")
	tag = strings.TrimSpace(tag)
	// Split by comma to handle options like omitempty
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return tag
}
