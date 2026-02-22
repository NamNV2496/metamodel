package generator

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func ParseFile(filename string, tag string) ([]StructMeta, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse file %s: %w", filename, err)
	}

	// Build local struct type map (same-file resolution)
	localStructTypes := collectStructTypes(node)

	// Build import map and module resolver for cross-package resolution
	imports := collectImports(node)
	modulePath, moduleRoot, _ := findModuleInfo(filename)
	resolver := &pkgResolver{
		imports:    imports,
		modulePath: modulePath,
		moduleRoot: moduleRoot,
		cache:      make(map[string]map[string]*ast.StructType),
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
				Fields:     parseFields(structType, tag, localStructTypes, resolver),
			}
			if len(meta.Fields) > 0 {
				structs = append(structs, meta)
			}
		}
	}

	return structs, packageName, nil
}

// collectStructTypes builds a map of struct name -> *ast.StructType for all structs in a file.
func collectStructTypes(node *ast.File) map[string]*ast.StructType {
	m := make(map[string]*ast.StructType)
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
			m[typeSpec.Name.Name] = structType
		}
	}
	return m
}

// collectImports builds a map of local package name -> import path.
func collectImports(node *ast.File) map[string]string {
	m := make(map[string]string)
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		var localName string
		if imp.Name != nil {
			localName = imp.Name.Name
		} else {
			// Use the last segment of the import path as the default package name
			parts := strings.Split(path, "/")
			localName = parts[len(parts)-1]
		}
		m[localName] = path
	}
	return m
}

// findModuleInfo walks up from the source file to find go.mod and reads the module path.
func findModuleInfo(sourceFile string) (modulePath, moduleRoot string, err error) {
	absFile, err := filepath.Abs(sourceFile)
	if err != nil {
		return "", "", err
	}
	dir := filepath.Dir(absFile)
	for {
		gomod := filepath.Join(dir, "go.mod")
		if _, statErr := os.Stat(gomod); statErr == nil {
			modulePath, err = readModulePath(gomod)
			if err != nil {
				return "", "", err
			}
			return modulePath, dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("go.mod not found")
}

// readModulePath reads the module directive from a go.mod file.
func readModulePath(gomodPath string) (string, error) {
	f, err := os.Open(gomodPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in %s", gomodPath)
}

// pkgResolver resolves struct types from external packages.
type pkgResolver struct {
	imports    map[string]string                     // local pkg name -> import path
	modulePath string                                // e.g., "github.com/namnv2496/metamodel"
	moduleRoot string                                // abs path to dir containing go.mod
	cache      map[string]map[string]*ast.StructType // import path -> struct name -> struct type
}

// resolveExternalStruct returns the *ast.StructType for pkgAlias.typeName, or nil if not found.
func (r *pkgResolver) resolveExternalStruct(pkgAlias, typeName string) (map[string]*ast.StructType, *ast.StructType) {
	importPath, ok := r.imports[pkgAlias]
	if !ok {
		return nil, nil
	}

	// Check cache
	if cached, ok := r.cache[importPath]; ok {
		return cached, cached[typeName]
	}

	// Compute the on-disk directory for this import path
	if r.moduleRoot == "" || r.modulePath == "" {
		return nil, nil
	}
	if !strings.HasPrefix(importPath, r.modulePath) {
		// External module — not resolvable without go/packages
		return nil, nil
	}
	relPath := strings.TrimPrefix(importPath, r.modulePath)
	relPath = strings.TrimPrefix(relPath, "/")
	pkgDir := filepath.Join(r.moduleRoot, relPath)

	// Parse all .go files in that directory
	structTypes := make(map[string]*ast.StructType)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, nil
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		filePath := filepath.Join(pkgDir, entry.Name())
		fileNode, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			continue
		}
		for name, st := range collectStructTypes(fileNode) {
			structTypes[name] = st
		}
	}

	r.cache[importPath] = structTypes
	return structTypes, structTypes[typeName]
}

func parseFields(structType *ast.StructType, tag string, structTypes map[string]*ast.StructType, resolver *pkgResolver) []FieldMeta {
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
				fields = append(fields, resolveEmbedded(field.Type, tag, structTypes, resolver)...)
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

func resolveEmbedded(fieldType ast.Expr, tag string, localStructTypes map[string]*ast.StructType, resolver *pkgResolver) []FieldMeta {
	switch t := fieldType.(type) {
	case *ast.Ident:
		// Same-package: Entity
		if st, ok := localStructTypes[t.Name]; ok {
			return parseFields(st, tag, localStructTypes, resolver)
		}
	case *ast.StarExpr:
		return resolveEmbedded(t.X, tag, localStructTypes, resolver)
	case *ast.SelectorExpr:
		// Cross-package: entity.Entity
		pkgIdent, ok := t.X.(*ast.Ident)
		if !ok {
			break
		}
		pkgAlias := pkgIdent.Name
		typeName := t.Sel.Name
		if resolver != nil {
			extStructTypes, st := resolver.resolveExternalStruct(pkgAlias, typeName)
			if st != nil {
				return parseFields(st, tag, extStructTypes, resolver)
			}
		}
	}
	return nil
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
