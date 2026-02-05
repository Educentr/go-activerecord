package postgres

import (
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/schema"
)

func TestPostgresSchemaGenerator_SupportsMigrations(t *testing.T) {
	g := NewSchemaGenerator()
	if !g.SupportsMigrations() {
		t.Error("PostgreSQL should support migrations")
	}
}

func TestPostgresSchemaGenerator_SchemaFileName(t *testing.T) {
	g := NewSchemaGenerator()
	if g.SchemaFileName() != "schema.sql" {
		t.Errorf("expected 'schema.sql', got '%s'", g.SchemaFileName())
	}
}

func TestPostgresSchemaGenerator_SchemaFileExtension(t *testing.T) {
	g := NewSchemaGenerator()
	if g.SchemaFileExtension() != ".sql" {
		t.Errorf("expected '.sql', got '%s'", g.SchemaFileExtension())
	}
}

func TestPostgresSchemaGenerator_GenerateSchemaJSON(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Namespace.PublicName = "users"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Id", Format: "int64", PrimaryKey: true},
		{Name: "Email", Format: "string", Size: 255},
		{Name: "Status", Format: "int32"},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
		{
			Name:      "Email",
			Fields:    []int{1},
			FieldsMap: map[string]ds.IndexField{"Email": {IndField: 1}},
			Unique:    true,
		},
	}

	table, err := g.GenerateSchemaJSON(pkg)
	if err != nil {
		t.Fatalf("GenerateSchemaJSON error: %v", err)
	}

	if table.Name != "users" {
		t.Errorf("expected table name 'users', got '%s'", table.Name)
	}
	if table.Backend != "postgres" {
		t.Errorf("expected backend 'postgres', got '%s'", table.Backend)
	}
	if len(table.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(table.Columns))
	}
	if len(table.Indexes) != 2 {
		t.Errorf("expected 2 indexes, got %d", len(table.Indexes))
	}
	if len(table.PrimaryKey) != 1 || table.PrimaryKey[0] != "id" {
		t.Errorf("expected primary key ['id'], got %v", table.PrimaryKey)
	}
}

func TestPostgresSchemaGenerator_GenerateFullSchema(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Namespace.PublicName = "users"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Id", Format: "int64", PrimaryKey: true},
		{Name: "Email", Format: "string", Size: 255},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	ddl, err := g.GenerateFullSchema(pkg)
	if err != nil {
		t.Fatalf("GenerateFullSchema error: %v", err)
	}

	ddlStr := string(ddl)

	if !strings.Contains(ddlStr, "CREATE TABLE IF NOT EXISTS users") {
		t.Error("expected DDL to contain CREATE TABLE statement")
	}
	if !strings.Contains(ddlStr, "id BIGINT NOT NULL DEFAULT 0") {
		t.Errorf("expected DDL to contain id column with NOT NULL and DEFAULT, got:\n%s", ddlStr)
	}
	if !strings.Contains(ddlStr, "email VARCHAR(255) NOT NULL DEFAULT ''") {
		t.Errorf("expected DDL to contain email column with NOT NULL and DEFAULT, got:\n%s", ddlStr)
	}
	if !strings.Contains(ddlStr, "PRIMARY KEY (id)") {
		t.Error("expected DDL to contain PRIMARY KEY constraint")
	}
}

func TestPostgresSchemaGenerator_GenerateFullSchema_WithObjectName(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Namespace.ObjectName = "app_user"
	pkg.Namespace.PublicName = "AppUser"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Id", Format: "int64", PrimaryKey: true, InitByDB: true},
		{Name: "Name", Format: "string", Size: 100},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	ddl, err := g.GenerateFullSchema(pkg)
	if err != nil {
		t.Fatalf("GenerateFullSchema error: %v", err)
	}

	ddlStr := string(ddl)

	// Проверяем использование snake_case имени из ObjectName
	if !strings.Contains(ddlStr, "CREATE TABLE IF NOT EXISTS app_user") {
		t.Errorf("expected DDL to use ObjectName 'app_user', got:\n%s", ddlStr)
	}
	// BIGSERIAL не должен иметь DEFAULT
	if !strings.Contains(ddlStr, "id BIGSERIAL NOT NULL,") {
		t.Errorf("expected id BIGSERIAL NOT NULL (without DEFAULT), got:\n%s", ddlStr)
	}
}

func TestPostgresSchemaGenerator_GenerateSchemaJSON_WithForeignKey(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Namespace.ObjectName = "subscription"
	pkg.Namespace.PublicName = "Subscription"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "UserID", Format: "int64", PrimaryKey: true},
		{Name: "RegCode", Format: "string", Size: 16},
	}
	pkg.FieldsObjectMap = map[string]ds.FieldObject{
		"User": {
			Name:       "User",
			Key:        "ID",
			ObjectName: "app_user",
			Field:      "UserID",
			Unique:     true,
		},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"UserID": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	table, err := g.GenerateSchemaJSON(pkg)
	if err != nil {
		t.Fatalf("GenerateSchemaJSON error: %v", err)
	}

	if table.Name != "subscription" {
		t.Errorf("expected table name 'subscription', got '%s'", table.Name)
	}

	if len(table.ForeignKeys) != 1 {
		t.Fatalf("expected 1 foreign key, got %d", len(table.ForeignKeys))
	}

	fk := table.ForeignKeys[0]
	if fk.Name != "subscription_userid_fkey" {
		t.Errorf("expected FK name 'subscription_userid_fkey', got '%s'", fk.Name)
	}
	if fk.Column != "userid" {
		t.Errorf("expected FK column 'userid', got '%s'", fk.Column)
	}
	if fk.RefTable != "app_user" {
		t.Errorf("expected FK ref table 'app_user', got '%s'", fk.RefTable)
	}
	if fk.RefColumn != "id" {
		t.Errorf("expected FK ref column 'id', got '%s'", fk.RefColumn)
	}
	if fk.OnDelete != "CASCADE" {
		t.Errorf("expected FK on delete 'CASCADE', got '%s'", fk.OnDelete)
	}
}

func TestPostgresSchemaGenerator_GenerateFullSchema_WithForeignKey(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Namespace.ObjectName = "subscription"
	pkg.Namespace.PublicName = "Subscription"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "UserID", Format: "int64", PrimaryKey: true},
		{Name: "RegCode", Format: "string", Size: 16},
	}
	pkg.FieldsObjectMap = map[string]ds.FieldObject{
		"User": {
			Name:       "User",
			Key:        "ID",
			ObjectName: "app_user",
			Field:      "UserID",
			Unique:     true,
		},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"UserID": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	ddl, err := g.GenerateFullSchema(pkg)
	if err != nil {
		t.Fatalf("GenerateFullSchema error: %v", err)
	}

	ddlStr := string(ddl)

	if !strings.Contains(ddlStr, "CONSTRAINT subscription_userid_fkey FOREIGN KEY (userid)") {
		t.Errorf("expected FK constraint, got:\n%s", ddlStr)
	}
	if !strings.Contains(ddlStr, "REFERENCES app_user (id) ON DELETE CASCADE") {
		t.Errorf("expected FK references, got:\n%s", ddlStr)
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_AddColumn(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{
		AddedColumns: []schema.Column{
			{Name: "Flags", Type: "BIGINT", Default: "0"},
		},
	}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration == nil {
		t.Fatal("expected non-nil migration")
	}
	if len(migration.Up) != 1 {
		t.Errorf("expected 1 UP statement, got %d", len(migration.Up))
	}
	if !strings.Contains(migration.Up[0], "ADD COLUMN Flags BIGINT") {
		t.Errorf("expected ADD COLUMN statement, got '%s'", migration.Up[0])
	}
	if len(migration.Down) != 1 {
		t.Errorf("expected 1 DOWN statement, got %d", len(migration.Down))
	}
	if !strings.Contains(migration.Down[0], "DROP COLUMN Flags") {
		t.Errorf("expected DROP COLUMN statement, got '%s'", migration.Down[0])
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_DropColumn(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{
		DroppedColumns: []string{"OldField"},
	}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration == nil {
		t.Fatal("expected non-nil migration")
	}
	if !strings.Contains(migration.Up[0], "DROP COLUMN OldField") {
		t.Errorf("expected DROP COLUMN statement, got '%s'", migration.Up[0])
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_AddIndex(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{
		AddedIndexes: []schema.Index{
			{Name: "idx_users_email", Columns: []string{"email"}, Unique: true},
		},
	}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration == nil {
		t.Fatal("expected non-nil migration")
	}
	if !strings.Contains(migration.Up[0], "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email") {
		t.Errorf("expected CREATE INDEX statement, got '%s'", migration.Up[0])
	}
	if !strings.Contains(migration.Down[0], "DROP INDEX IF EXISTS idx_users_email") {
		t.Errorf("expected DROP INDEX statement, got '%s'", migration.Down[0])
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_Empty(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration != nil {
		t.Error("expected nil migration for empty diff")
	}
}

func TestPostgresSchemaGenerator_goTypeToSQLType(t *testing.T) {
	g := NewSchemaGenerator()

	tests := []struct {
		format   ds.Format
		size     int64
		initByDB bool
		want     string
	}{
		{"int64", 0, false, "BIGINT"},
		{"int32", 0, false, "INTEGER"},
		{"int16", 0, false, "SMALLINT"},
		{"float64", 0, false, "DOUBLE PRECISION"},
		{"float32", 0, false, "REAL"},
		{"bool", 0, false, "BOOLEAN"},
		{"string", 255, false, "VARCHAR(255)"},
		{"string", 0, false, "TEXT"},
		{"time.Time", 0, false, "TIMESTAMP"},
		// Serial types (init_by_db = true)
		{"int64", 0, true, "BIGSERIAL"},
		{"int32", 0, true, "SERIAL"},
		{"int16", 0, true, "SMALLSERIAL"},
	}

	for _, tt := range tests {
		got := g.goTypeToSQLType(tt.format, tt.size, tt.initByDB)
		if got != tt.want {
			t.Errorf("goTypeToSQLType(%s, %d, %v) = %s, want %s", tt.format, tt.size, tt.initByDB, got, tt.want)
		}
	}
}

func TestPostgresSchemaGenerator_getDefaultValue(t *testing.T) {
	g := NewSchemaGenerator()

	tests := []struct {
		format ds.Format
		want   string
	}{
		{"int64", "0"},
		{"int32", "0"},
		{"int16", "0"},
		{"int8", "0"},
		{"int", "0"},
		{"uint64", "0"},
		{"uint32", "0"},
		{"uint16", "0"},
		{"uint8", "0"},
		{"uint", "0"},
		{"float64", "0"},
		{"float32", "0"},
		{"bool", "false"},
		{"string", "''"},
		{"[]byte", "''"},
		{"time.Time", "'0001-01-01 00:00:00'"}, // Значение time.Time{}
		{"CustomType", ""},                      // Неизвестные типы без DEFAULT
	}

	for _, tt := range tests {
		got := g.getDefaultValue(tt.format)
		if got != tt.want {
			t.Errorf("getDefaultValue(%s) = %q, want %q", tt.format, got, tt.want)
		}
	}
}
