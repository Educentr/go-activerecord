# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `go-activerecord` (v3), an Active Record ORM implementation for Go that generates type-safe database access code from declarative model definitions. It supports multiple database backends (PostgreSQL, Octopus, Tarantool 1.5).  and generates type-safe model packages with CRUD operations, selectors, and mutators.

The project consists of two main components:

1. **argen** - A code generator (`cmd/argen`) that transforms declarative model definitions into complete repository implementations
2. **Runtime packages** - Libraries in `pkg/` that the generated code depends on for database operations

## Key Commands

### Build

```bash
make build          # Build argen binary (output: bin/argen)
make install        # Install argen to $GOPATH/bin
```

### Testing

```bash
make test                    # Run all tests with 30s timeout
make cover                   # Generate coverage report and open in browser
TEST_TIMEOUT=60s make test   # Run tests with custom timeout
```

### Code Quality

```bash
make lint           # Run golangci-lint on changes from origin/main
make full-lint      # Run golangci-lint on entire codebase
make lint-fix       # Run linter and auto-fix issues
make generate       # Run go generate (required before commits)
```

### Development Tools

```bash
make install-tool        # Install mockery for generating mocks
make pre-commit-hook     # Install pre-commit hook (runs generate + lint)
make pre-push-hook       # Install pre-push hook (runs coverage tests)
```

### Running argen

```bash
# Basic usage - generate models from declarations
argen --path "model/repository" --declaration "decl" --destination "cmpl"

# With fixtures for testing
argen --path "model/repository" --declaration "decl" --destination "cmpl" \
      --fixture_path "testutil/fixture"

# Specify module name explicitly
argen --path "model/repository" --module "github.com/myorg/myapp"
```

## Architecture

### Three-Phase Generation Process

The argen tool operates in three phases:

1. **Parsing** (`internal/pkg/parser/`) - Reads Go files with special AR comments and struct tags to extract model definitions
2. **Validation** (`internal/pkg/checker/`) - Validates model declarations for correctness and consistency
3. **Code Generation** (`internal/pkg/generator/`) - Generates complete packages with CRUD methods, selectors, and backend-specific code

### Core Package Structure

```
pkg/
├── activerecord/     # Core interfaces, cluster management, connection pooling
├── octopus/          # Octopus/Tarantool 1.5 backend implementation
├── postgres/         # PostgreSQL backend implementation
├── iproto/           # Low-level iproto protocol implementation
├── serializer/       # Built-in serializers (Json, Printf, Mapstructure)
└── logger/           # Logging abstractions

internal/
├── app/              # Main argen application logic
└── pkg/
    ├── parser/       # Parses declarative model definitions from Go source
    ├── checker/      # Validates parsed models
    ├── generator/    # Code generation engine with templates
    └── backend/      # Backend-specific implementations (octopus, postgres)

cmd/argen/            # CLI entry point for code generator
```

### Model Declaration System

Models are declared using special comments (`//ar:`) and struct tags. Main struct types:

- **Fields\*** - Database fields/columns with tags for primary keys, indexes, serializers, mutators
- **Indexes\*** - Multi-column indexes with uniqueness, ordering, conditional clauses
- **IndexParts\*** - Partial index selectors for querying by subset of multi-column index
- **Serializers\*** - Custom serialization for complex field types
- **Mutators\*** - Atomic database operations (increment, bit operations, custom procedures)
- **Triggers\*** - Handlers for exceptional cases (e.g., tuple repair)
- **Flags\*** - Named bit flags for integer fields (generates constants and auto-adds set_bit/clear_bit mutators)
- **ProcFields\*** - For stored procedure/function calls (octopus backend)

### Generated Code Features

For each model declaration, argen generates:

- **Struct** with all fields plus metadata (Exists, IsReplica, Readonly, UpdateOps, Repaired)
- **CRUD methods**: Insert, Update, Replace, InsertOrReplace, Delete
- **Selectors**: SelectBy{Index} for each index (singular and plural variants)
- **Accessors**: Get{Field}/Set{Field} pairs with primary key protection
- **Mutators**: Inc{Field}, Dec{Field}, SetBit{Field}, ClearBit{Field}, And/Or/Xor{Field}
- **Flag constants**: {FieldName}{FlagName}Flag for each named flag (e.g., `FlagsActiveFlag = 1 << 0`)
- **Backend-specific code**: Connection handling, query building, serialization

### Configuration Architecture

The system uses a hierarchical configuration structure:

- **Cluster level**: Timeout, PoolSize for entire cluster
- **Shard level**: Per-shard timeouts, pool sizes, and server lists
- **Server level**: master (read-write) and replica (read-only) server lists

Configuration is accessed via the `ConfigInterface`, allowing integration with any config system (onlineconf, static, etc.).

### Backend Implementations

**Octopus/Tarantool 1.5**:

- Uses custom iproto binary protocol
- Namespaces identified by numeric ID
- Field order matters (must match tuple order)
- Supports stored procedures via ProcFields
- Extra fields beyond declaration stored in extraFields

**PostgreSQL**:

- Uses pgx/v5 driver
- Table names specified via `//ar:namespace:table_name` comment
- Supports conditional indexes (WHERE clauses)
- Full SQL query building with prepared statements

## Important Development Notes

### Code Generation Best Practices

- **Always run `make generate` before committing** - required for interface mocks
- **Package names must match regex `^[a-z]{1,20}$`** - lowercase letters only, max 20 chars
- **Field order is critical for Octopus** - must match tuple field order exactly
- **Generated files are deleted on regeneration** - never manually edit files in destination directory
- **The `.argen` marker file** in destination directory prevents accidental generation to wrong location

### Testing Considerations

- Use `//go:build activerecord` tag for tests that require generated code
- Fixture system available for integration tests (via --fixture_path)
- Run tests with `-tags=activerecord` build flag
- Mock servers available in pkg/octopus/mock_server.go for testing without real database

### Common Patterns

**Creating new records**:

```go
obj := model.New(ctx)
obj.SetField1(value)
obj.Insert(ctx)
```

**Querying records**:

```go
// Single record by unique index
obj, err := model.SelectById(ctx, id)

// Multiple records by non-unique index with limit
objs, err := model.SelectByType(ctx, typeVal, activerecord.NewLimiter(100))
```

**Atomic updates**:

```go
obj.IncCounter(10)                       // Increments counter by 10
obj.SetBitFlags(model.FlagsPremiumFlag)  // Sets bit flag using named constant
obj.Update(ctx)                          // Sends atomic operations to database
```

**Working with serializers**:

```go
// Field with Json serializer appears as deserialized type
jsonData := obj.GetJsonField()  // Returns map[string]interface{}
obj.SetJsonField(newData)       // Accepts map[string]interface{}
```

### Metrics and Observability

The generated code collects metrics via `MetricInterface`:

- Timing metrics: select_db, update_db, delete_db, insertreplace_db, call_proc
- Statistical metrics: insert_success, update_success, select_tuples_res
- Error metrics: select_db, update_db, delete_db with error suffixes

Metrics are namespace-specific and include operation details.

## Language and Documentation

- Main documentation is in **Russian** (README.md, docs/)
- Code comments are primarily in Russian
- Generated code and public APIs use English naming conventions
- Test reference implementation: [activerecord-cookbook](https://github.com/mailru/activerecord-cookbook)

## Minimum Requirements

- **Go 1.19.0+** (checked by scripts/goversioncheck.sh)
- **golangci-lint 1.60.3** (auto-installed by make targets)

## Common Development Tasks

### Adding a new backend

1. Create directory in `internal/pkg/backend/{newbackend}/`
2. Implement `backend.Backend` interface
3. Create checker implementing validation
4. Add templates in `tmpl/pkg/`
5. Register backend in `internal/pkg/backend/backend.go`

### Modifying generated code

1. Edit templates in `internal/pkg/backend/{backend}/tmpl/`
2. Test with `make test`
3. Regenerate test fixtures if needed

### Adding a field tag option

1. Update parser in `internal/pkg/parser/field.go`
2. Update `ds.Field` in `internal/pkg/ds/`
3. Add validation in appropriate backend checker
4. Update templates to use the new option

### Working with templates

- Templates use Go's text/template package
- Template functions are defined in `internal/pkg/generator/template.go`
- The primary data structure passed to templates is `*ds.RecordPackage`

## Important Notes

- Generated code is identified by file patterns `*_gen.go` and `*_mock.go` (excluded from linting)
- The project requires Go 1.19.0+ (enforced by `scripts/goversioncheck.sh`)
- Tests use build tag `activerecord` - always run with this tag
- Field order matters for Octopus backend (must match tuple structure)
- Index order matters for Octopus backend (must match database configuration)

## Release Process

Steps for creating a new release:

1. **Verify all changes are committed**
   ```bash
   git status  # Should show clean working tree
   ```

2. **Run linters and verify no new issues since last release**
   ```bash
   make lint      # Check changes from origin/main
   make full-lint # Full codebase check
   ```

3. **Check test coverage**
   ```bash
   make cover
   ```
   Coverage should not decrease compared to previous release.

4. **Run all tests**
   ```bash
   make test
   ```
   All tests must pass.

5. **Update documentation**
   - Update docs/ for all changes made
   - Update README.md if public API changed
   - Update CLAUDE.md if development workflow changed

6. **Create and push tag**
   ```bash
   git tag v3.X.Y
   git push origin v3.X.Y
   ```
