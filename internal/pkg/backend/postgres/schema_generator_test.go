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

func TestPostgresSchemaGenerator_ConditionalIndex(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Namespace.ObjectName = "sessions"
	pkg.Namespace.PublicName = "Session"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "ID", Format: "int64", PrimaryKey: true, InitByDB: true},
		{Name: "Status", Format: "string", Size: 32},
		{Name: "UserID", Format: "int64"},
		{Name: "LastKeepAlive", Format: "time.Time"},
	}
	pkg.FieldsMap = map[string]int{
		"ID": 0, "Status": 1, "UserID": 2, "LastKeepAlive": 3,
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"ID": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
		{
			Name:      "ActiveUserByKeepAlive",
			Fields:    []int{2, 3},
			FieldsMap: map[string]ds.IndexField{"UserID": {IndField: 2}, "LastKeepAlive": {IndField: 3}},
			Conditions: map[int]ds.IndexCondition{
				1: {ConditionType: "=", Value: []string{"active"}},
			},
		},
	}

	ddl, err := g.GenerateFullSchema(pkg)
	if err != nil {
		t.Fatalf("GenerateFullSchema error: %v", err)
	}

	ddlStr := string(ddl)

	if !strings.Contains(ddlStr, "idx_sessions_activeuserbykeepalive") {
		t.Errorf("expected conditional index in DDL, got:\n%s", ddlStr)
	}

	if !strings.Contains(ddlStr, "WHERE status = 'active'") {
		t.Errorf("expected WHERE clause in DDL, got:\n%s", ddlStr)
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
		{"CustomType", ""},                     // Неизвестные типы без DEFAULT
	}

	for _, tt := range tests {
		got := g.getDefaultValue(tt.format)
		if got != tt.want {
			t.Errorf("getDefaultValue(%s) = %q, want %q", tt.format, got, tt.want)
		}
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_NilDiff(t *testing.T) {
	g := NewSchemaGenerator()

	migration, err := g.GenerateMigration(nil, "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if migration != nil {
		t.Errorf("expected nil migration for nil diff, got %+v", migration)
	}
}

func TestPostgresSchemaGenerator_GenerateSchemaJSON_WrongType(t *testing.T) {
	g := NewSchemaGenerator()

	_, err := g.GenerateSchemaJSON("not a RecordPackage")
	if err == nil {
		t.Fatal("expected error for wrong type")
	}

	if !strings.Contains(err.Error(), "expected *ds.RecordPackage") {
		t.Errorf("error %q should contain 'expected *ds.RecordPackage'", err.Error())
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_ModifiedColumn(t *testing.T) {
	g := NewSchemaGenerator()

	tests := []struct {
		name       string
		diff       *schema.TableDiff
		upContains []string
		dnContains []string
	}{
		{
			name: "type changed",
			diff: &schema.TableDiff{
				ModifiedColumns: map[string]schema.ColumnDiff{
					"status": {
						OldColumn:   schema.Column{Name: "status", Type: "INTEGER"},
						NewColumn:   schema.Column{Name: "status", Type: "BIGINT"},
						TypeChanged: true,
					},
				},
			},
			upContains: []string{"ALTER COLUMN status TYPE BIGINT"},
			dnContains: []string{"ALTER COLUMN status TYPE INTEGER"},
		},
		{
			name: "not null set",
			diff: &schema.TableDiff{
				ModifiedColumns: map[string]schema.ColumnDiff{
					"email": {
						OldColumn:      schema.Column{Name: "email", NotNull: false},
						NewColumn:      schema.Column{Name: "email", NotNull: true},
						NotNullChanged: true,
					},
				},
			},
			upContains: []string{"SET NOT NULL"},
			dnContains: []string{"DROP NOT NULL"},
		},
		{
			name: "not null dropped",
			diff: &schema.TableDiff{
				ModifiedColumns: map[string]schema.ColumnDiff{
					"email": {
						OldColumn:      schema.Column{Name: "email", NotNull: true},
						NewColumn:      schema.Column{Name: "email", NotNull: false},
						NotNullChanged: true,
					},
				},
			},
			upContains: []string{"DROP NOT NULL"},
			dnContains: []string{"SET NOT NULL"},
		},
		{
			name: "default changed",
			diff: &schema.TableDiff{
				ModifiedColumns: map[string]schema.ColumnDiff{
					"count": {
						OldColumn:      schema.Column{Name: "count", Default: "0"},
						NewColumn:      schema.Column{Name: "count", Default: "42"},
						DefaultChanged: true,
					},
				},
			},
			upContains: []string{"SET DEFAULT 42"},
			dnContains: []string{"SET DEFAULT 0"},
		},
		{
			name: "default dropped",
			diff: &schema.TableDiff{
				ModifiedColumns: map[string]schema.ColumnDiff{
					"count": {
						OldColumn:      schema.Column{Name: "count", Default: "0"},
						NewColumn:      schema.Column{Name: "count", Default: ""},
						DefaultChanged: true,
					},
				},
			},
			upContains: []string{"DROP DEFAULT"},
			dnContains: []string{"SET DEFAULT 0"},
		},
		{
			name: "default added",
			diff: &schema.TableDiff{
				ModifiedColumns: map[string]schema.ColumnDiff{
					"count": {
						OldColumn:      schema.Column{Name: "count", Default: ""},
						NewColumn:      schema.Column{Name: "count", Default: "99"},
						DefaultChanged: true,
					},
				},
			},
			upContains: []string{"SET DEFAULT 99"},
			dnContains: []string{"DROP DEFAULT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migration, err := g.GenerateMigration(tt.diff, "users")
			if err != nil {
				t.Fatalf("GenerateMigration error: %v", err)
			}

			if migration == nil {
				t.Fatal("expected non-nil migration")
			}

			upJoined := strings.Join(migration.Up, "\n")

			for _, substr := range tt.upContains {
				if !strings.Contains(upJoined, substr) {
					t.Errorf("Up statements should contain %q, got:\n%s", substr, upJoined)
				}
			}

			dnJoined := strings.Join(migration.Down, "\n")

			for _, substr := range tt.dnContains {
				if !strings.Contains(dnJoined, substr) {
					t.Errorf("Down statements should contain %q, got:\n%s", substr, dnJoined)
				}
			}
		})
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_ModifiedIndex(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{
		ModifiedIndexes: map[string]schema.IndexDiff{
			"idx_email": {
				OldIndex:      schema.Index{Name: "idx_email", Columns: []string{"email"}, Unique: false},
				NewIndex:      schema.Index{Name: "idx_email", Columns: []string{"email"}, Unique: true},
				UniqueChanged: true,
			},
		},
	}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration == nil {
		t.Fatal("expected non-nil migration")
	}

	upJoined := strings.Join(migration.Up, "\n")

	if !strings.Contains(upJoined, "DROP INDEX IF EXISTS idx_email") {
		t.Errorf("Up should contain DROP INDEX, got:\n%s", upJoined)
	}

	if !strings.Contains(upJoined, "CREATE UNIQUE INDEX IF NOT EXISTS idx_email") {
		t.Errorf("Up should contain CREATE UNIQUE INDEX, got:\n%s", upJoined)
	}

	dnJoined := strings.Join(migration.Down, "\n")

	if !strings.Contains(dnJoined, "DROP INDEX IF EXISTS idx_email") {
		t.Errorf("Down should contain DROP INDEX, got:\n%s", dnJoined)
	}

	// Down should recreate without UNIQUE
	if strings.Contains(dnJoined, "CREATE UNIQUE INDEX") {
		t.Errorf("Down should NOT contain CREATE UNIQUE INDEX (original was not unique), got:\n%s", dnJoined)
	}

	if !strings.Contains(dnJoined, "CREATE INDEX IF NOT EXISTS idx_email") {
		t.Errorf("Down should contain CREATE INDEX (non-unique), got:\n%s", dnJoined)
	}
}

func TestPostgresSchemaGenerator_GenerateMigration_DropIndex(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{
		DroppedIndexes: []string{"idx_old"},
	}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration == nil {
		t.Fatal("expected non-nil migration")
	}

	upJoined := strings.Join(migration.Up, "\n")

	if !strings.Contains(upJoined, "DROP INDEX IF EXISTS idx_old") {
		t.Errorf("Up should contain DROP INDEX IF EXISTS idx_old, got:\n%s", upJoined)
	}
}

func TestPostgresSchemaGenerator_generateCreateIndexStatement(t *testing.T) {
	g := NewSchemaGenerator()

	tests := []struct {
		name     string
		table    string
		idx      schema.Index
		contains []string
	}{
		{
			name:  "simple",
			table: "users",
			idx:   schema.Index{Name: "idx_users_email", Columns: []string{"email"}},
			contains: []string{
				"CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)",
			},
		},
		{
			name:  "unique",
			table: "users",
			idx:   schema.Index{Name: "idx_users_email", Columns: []string{"email"}, Unique: true},
			contains: []string{
				"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email",
			},
		},
		{
			name:  "with order",
			table: "users",
			idx: schema.Index{
				Name:    "idx_users_email",
				Columns: []string{"email"},
				Order:   map[string]string{"email": "DESC"},
			},
			contains: []string{"email DESC"},
		},
		{
			name:  "with condition",
			table: "users",
			idx: schema.Index{
				Name:      "idx_users_active",
				Columns:   []string{"status"},
				Condition: "status = 'active'",
			},
			contains: []string{"WHERE status = 'active'"},
		},
		{
			name:  "multi-column mixed order",
			table: "events",
			idx: schema.Index{
				Name:    "idx_events_compound",
				Columns: []string{"user_id", "created_at"},
				Order:   map[string]string{"user_id": "ASC", "created_at": "DESC"},
			},
			contains: []string{"user_id ASC", "created_at DESC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.generateCreateIndexStatement(tt.table, tt.idx)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("result should contain %q, got: %s", substr, result)
				}
			}
		})
	}
}

func TestPostgresSchemaGenerator_generateConditionClause(t *testing.T) {
	g := NewSchemaGenerator()

	tests := []struct {
		name       string
		conditions map[int]ds.IndexCondition
		pkg        *ds.RecordPackage
		want       string
	}{
		{
			name:       "empty conditions",
			conditions: map[int]ds.IndexCondition{},
			pkg: func() *ds.RecordPackage {
				p := ds.NewRecordPackage()
				return p
			}(),
			want: "",
		},
		{
			name: "simple equality with string",
			conditions: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"active"}},
			},
			pkg: func() *ds.RecordPackage {
				p := ds.NewRecordPackage()
				p.Fields = []ds.FieldDeclaration{
					{Name: "Status", Format: "string", Size: 32},
				}
				return p
			}(),
			want: "status = 'active'",
		},
		{
			name: "IS NULL check",
			conditions: map[int]ds.IndexCondition{
				0: {ConditionType: "is null", IsNullCheck: true},
			},
			pkg: func() *ds.RecordPackage {
				p := ds.NewRecordPackage()
				p.Fields = []ds.FieldDeclaration{
					{Name: "Status", Format: "string", Size: 32},
				}
				return p
			}(),
			want: "status is null",
		},
		{
			name: "field expression bitwise",
			conditions: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"1"}, FieldExpression: "Flags&1"},
			},
			pkg: func() *ds.RecordPackage {
				p := ds.NewRecordPackage()
				p.Fields = []ds.FieldDeclaration{
					{Name: "Flags", Format: "int64"},
				}
				return p
			}(),
			want: "flags&1 = 1",
		},
		{
			name: "fieldNum out of range",
			conditions: map[int]ds.IndexCondition{
				99: {ConditionType: "=", Value: []string{"x"}},
			},
			pkg: func() *ds.RecordPackage {
				p := ds.NewRecordPackage()
				p.Fields = []ds.FieldDeclaration{
					{Name: "ID", Format: "int64"},
				}
				return p
			}(),
			want: "",
		},
		{
			name: "integer not quoted",
			conditions: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"42"}},
			},
			pkg: func() *ds.RecordPackage {
				p := ds.NewRecordPackage()
				p.Fields = []ds.FieldDeclaration{
					{Name: "Id", Format: "int64"},
				}
				return p
			}(),
			want: "id = 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.generateConditionClause(tt.conditions, tt.pkg)
			if got != tt.want {
				t.Errorf("generateConditionClause() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPostgresSchemaGenerator_generateConditionClause_MultipleConditions(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Status", Format: "string", Size: 32},
		{Name: "Active", Format: "bool"},
	}

	// Use ordered conditions - note: map iteration order is non-deterministic,
	// so we check both parts are present
	conditions := map[int]ds.IndexCondition{
		0: {ConditionType: "=", Value: []string{"published"}},
		1: {ConditionType: "=", Value: []string{"true"}},
	}

	got := g.generateConditionClause(conditions, pkg)

	if !strings.Contains(got, "status = 'published'") {
		t.Errorf("result should contain status condition, got: %q", got)
	}

	if !strings.Contains(got, "active = true") {
		t.Errorf("result should contain active condition, got: %q", got)
	}

	if !strings.Contains(got, " AND ") {
		t.Errorf("multiple conditions should be joined with AND, got: %q", got)
	}
}
