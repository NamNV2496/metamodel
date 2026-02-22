# Metamodel - Go Code Generator for Struct Field Constants

In development the hard code is very popular. Especially, when work with database example when query "select * from table where 'id'=xxx and 'name'='yyy'".

Metamodel is a code generation tool that scans Go structs with `json` or `bson` tags and generates type-safe field name constants. This helps eliminate string literals in your code and provides compile-time safety when working with struct field names.

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

### How to remove

```bash
rm $(go env GOPATH)/bin/metamodel
```
## Usage

### Using with go generate (Recommended)

Add a `//go:generate` directive to your Go file:

```go
package repository

//go:generate metamodel -source=$GOFILE -destination=../generated/
type Scenarios struct {
	FeatureName string `json:"feature_name,omitempty"`
	ScenarioID  int    `json:"scenario_id"`
	Status      string `bson:"status"`
	Description string `json:"description,omitempty" bson:"desc"`
	IgnoreMe    string // no tag, should be ignored
	SkippedTag  string `json:"-"` // skip tag
}
```

Then run:

```bash
go generate ./...
```

### Direct Command Line Usage

```bash
metamodel -source=path/to/your/file.go -destination=path/to/generate_file.go -tag=bson
```

# Example

```go
package main

import (
	"fmt"

	repository_ "github.com/namnv2496/exmaple/generated"
)

func main() {
	// Use the generated metamodel constants
	fmt.Println("Scenarios.TableName:", repository_.Scenarios_.TableName)
	fmt.Println("Scenarios.FeatureName:", repository_.Scenarios_.FeatureName)
	fmt.Println("Scenarios.Status:", repository_.Scenarios_.Status)
	fmt.Println("Feature.FeatureName:", repository_.Feature_.FeatureName)
	fmt.Println("AnotherModel.UserID:", repository_.AnotherModel_.UserID)
}
```