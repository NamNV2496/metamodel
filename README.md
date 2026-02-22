# Metamodel - Go Code Generator for Struct Field Constants

Metamodel is a code generation tool that scans Go structs with `json` or `bson` tags and generates type-safe field name constants. This helps eliminate string literals in your code and provides compile-time safety when working with struct field names.

## Features

- ✅ Scans `json` and `bson` struct tags
- ✅ Generates type-safe field name constants
- ✅ Works with `go generate` directive
- ✅ Installable via `go install`
- ✅ Supports selective type generation
- ✅ Skips fields without tags or with `-` tag

## Installation

Install the tool using `go install`:

```bash
go install github.com/namnv2496/metamodel@latest
```

Or build from source:

```bash
git clone https://github.com/namnv2496/metamodel.git
cd metamodel
go install .
```

## Usage

### Using with go generate (Recommended)

Add a `//go:generate` directive to your Go file:

```go
package repository

//go:generate go run github.com/namnv2496/metamodel -source=scenarios.go

type Scenarios struct {
    FeatureName string `json:"feature_name,omitempty"`
    ScenarioID  int    `json:"scenario_id"`
    Status      string `bson:"status"`
    Description string `json:"description,omitempty" bson:"desc"`
}
```

Then run:

```bash
go generate ./...
```

### Direct Command Line Usage

```bash
metamodel -source=path/to/your/file.go
```

#### Available Flags

- `-source` (required): Source file to generate metamodel from
- `-destination`: Output file for generated code (default: `<source>_metamodel.go`)
- `-package`: Package name for generated file (default: same as source)
- `-tag`: Specific tag name to generate (optional, eg: json, bson generates all if not specified)

#### Examples

Generate for all structs in a file:
```bash
metamodel -source=models.go
```

Generate for a specific tag:
```bash
metamodel -source=models.go -tag=bson
```

Specify custom output:
```bash
metamodel -source=models.go -destination=generated_models.go
```
