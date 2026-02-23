package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/namnv2496/metamodel/generator"
)

var (
	source      = flag.String("source", "", "Source file to generate metamodel from (e.g., models.go)")
	destination = flag.String("destination", "", "Output file for generated code (default: <source>_metamodel.go, e.g., models_metamodel.go)")
	packageName = flag.String("packageName", "", "Package name for generated file (default: same as source, e.g., models)")
	tag         = flag.String("tag", "json", "Specific tag name to generate (optional, e.g., json, bson, gorm)")
	tableName   = flag.String("tableName", "", "Specific table name to generate (optional, e.g., json, bson, gorm)")
)

func main() {
	flag.Parse()

	if *source == "" {
		fmt.Fprintln(os.Stderr, "Error: -source flag is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg := generator.Config{
		Source:      *source,
		Destination: *destination,
		PackageName: *packageName,
		Tag:         *tag,
		TableName:   *tableName,
	}

	if err := generator.Generate(cfg); err != nil {
		log.Fatalf("Error generating metamodel: %v", err)
	}

	fmt.Printf("Successfully generated metamodel for %s\n", *source)
}
