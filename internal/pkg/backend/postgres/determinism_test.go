package postgres

import (
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/schema"
)

// Порядок частей WHERE задаётся номером поля: иначе строка condition в
// schema.json меняется между прогонами, computeIndexDiff видит ложное
// изменение индекса и создаёт лишнюю миграцию с DROP/CREATE INDEX
func TestPostgresSchemaGenerator_generateConditionClause_Deterministic(t *testing.T) {
	g := NewSchemaGenerator()

	pkg := ds.NewRecordPackage()
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Status", Format: "string", Size: 32},
		{Name: "Flags", Format: "uint32"},
		{Name: "DeletedAt", Format: "int64"},
		{Name: "Kind", Format: "string", Size: 16},
	}

	conditions := map[int]ds.IndexCondition{
		3: {ConditionType: "=", Value: []string{"user"}},
		0: {ConditionType: "=", Value: []string{"published"}},
		2: {ConditionType: "is null", IsNullCheck: true},
		1: {ConditionType: "=", Value: []string{"1"}, FieldExpression: "Flags&1"},
	}

	want := "status = 'published' AND flags&1 = 1 AND deletedat is null AND kind = 'user'"

	for i := 0; i < 20; i++ {
		if got := g.generateConditionClause(conditions, pkg); got != want {
			t.Fatalf("iteration %d: generateConditionClause() = %q, want %q", i, got, want)
		}
	}
}

// Порядок CONSTRAINT в DDL (и foreign_keys в schema.json) не должен зависеть от
// обхода FieldsObjectMap
func TestPostgresSchemaGenerator_ForeignKeysDeterministic(t *testing.T) {
	g := NewSchemaGenerator()

	newPkg := func() *ds.RecordPackage {
		pkg := ds.NewRecordPackage()
		pkg.Namespace.ObjectName = "subscription"
		pkg.Namespace.PublicName = "Subscription"
		pkg.Fields = []ds.FieldDeclaration{
			{Name: "ID", Format: "int64", PrimaryKey: true},
			{Name: "UserID", Format: "int64"},
			{Name: "AccountID", Format: "int64"},
			{Name: "ProductID", Format: "int64"},
		}
		pkg.FieldsObjectMap = map[string]ds.FieldObject{
			"User":    {Name: "User", Key: "ID", ObjectName: "app_user", Field: "UserID", Unique: false},
			"Account": {Name: "Account", Key: "ID", ObjectName: "account", Field: "AccountID", Unique: false},
			"Product": {Name: "Product", Key: "ID", ObjectName: "product", Field: "ProductID", Unique: false},
		}
		pkg.Indexes = []ds.IndexDeclaration{
			{
				Name:      "PK",
				Fields:    []int{0},
				FieldsMap: map[string]ds.IndexField{"ID": {IndField: 0}},
				Primary:   true,
				Unique:    true,
			},
		}

		return pkg
	}

	// Порядок — по имени поля связи: Account, Product, User
	wantFKs := []string{
		"subscription_accountid_fkey",
		"subscription_productid_fkey",
		"subscription_userid_fkey",
	}

	for i := 0; i < 20; i++ {
		table, err := g.GenerateSchemaJSON(newPkg())
		if err != nil {
			t.Fatalf("GenerateSchemaJSON error: %v", err)
		}

		if len(table.ForeignKeys) != len(wantFKs) {
			t.Fatalf("expected %d foreign keys, got %d", len(wantFKs), len(table.ForeignKeys))
		}

		for j, want := range wantFKs {
			if table.ForeignKeys[j].Name != want {
				t.Fatalf("iteration %d: ForeignKeys[%d] = %s, want %s", i, j, table.ForeignKeys[j].Name, want)
			}
		}
	}
}

// Операторы миграции для изменённых колонок и индексов идут в порядке имён
func TestPostgresSchemaGenerator_GenerateMigration_Deterministic(t *testing.T) {
	g := NewSchemaGenerator()

	diff := &schema.TableDiff{
		ModifiedColumns: map[string]schema.ColumnDiff{
			"zipcode": {
				OldColumn:   schema.Column{Name: "zipcode", Type: "VARCHAR(8)"},
				NewColumn:   schema.Column{Name: "zipcode", Type: "VARCHAR(16)"},
				TypeChanged: true,
			},
			"email": {
				OldColumn:   schema.Column{Name: "email", Type: "VARCHAR(64)"},
				NewColumn:   schema.Column{Name: "email", Type: "VARCHAR(256)"},
				TypeChanged: true,
			},
		},
		ModifiedIndexes: map[string]schema.IndexDiff{
			"idx_users_status": {
				OldIndex:       schema.Index{Name: "idx_users_status", Columns: []string{"status"}},
				NewIndex:       schema.Index{Name: "idx_users_status", Columns: []string{"status", "created_at"}},
				ColumnsChanged: true,
			},
			"idx_users_email": {
				OldIndex:      schema.Index{Name: "idx_users_email", Columns: []string{"email"}},
				NewIndex:      schema.Index{Name: "idx_users_email", Columns: []string{"email"}, Unique: true},
				UniqueChanged: true,
			},
		},
	}

	wantUp := []string{
		"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(256)",
		"ALTER TABLE users ALTER COLUMN zipcode TYPE VARCHAR(16)",
		"DROP INDEX IF EXISTS idx_users_email",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email)",
		"DROP INDEX IF EXISTS idx_users_status",
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users (status, created_at)",
	}

	wantDescription := "change type of email, change type of zipcode, modify index idx_users_email, modify index idx_users_status"

	for i := 0; i < 20; i++ {
		migration, err := g.GenerateMigration(diff, "users")
		if err != nil {
			t.Fatalf("GenerateMigration error: %v", err)
		}

		if strings.Join(migration.Up, "\n") != strings.Join(wantUp, "\n") {
			t.Fatalf("iteration %d: migration.Up =\n%s\nwant:\n%s", i, strings.Join(migration.Up, "\n"), strings.Join(wantUp, "\n"))
		}

		if migration.Description != wantDescription {
			t.Fatalf("iteration %d: migration.Description = %q, want %q", i, migration.Description, wantDescription)
		}
	}
}

// Полная DDL схема таблицы побайтно одинакова между прогонами
func TestPostgresSchemaGenerator_GenerateFullSchema_Deterministic(t *testing.T) {
	g := NewSchemaGenerator()

	newPkg := func() *ds.RecordPackage {
		pkg := ds.NewRecordPackage()
		pkg.Namespace.ObjectName = "users"
		pkg.Namespace.PublicName = "Users"
		pkg.Fields = []ds.FieldDeclaration{
			{Name: "ID", Format: "int64", PrimaryKey: true, InitByDB: true},
			{Name: "Email", Format: "string", Size: 256},
			{Name: "Status", Format: "string", Size: 32},
			{Name: "Flags", Format: "uint32"},
			{Name: "AccountID", Format: "int64"},
			{Name: "ProductID", Format: "int64"},
		}
		pkg.FieldsObjectMap = map[string]ds.FieldObject{
			"Account": {Name: "Account", Key: "ID", ObjectName: "account", Field: "AccountID"},
			"Product": {Name: "Product", Key: "ID", ObjectName: "product", Field: "ProductID"},
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
				Name:      "Active",
				Fields:    []int{1},
				FieldsMap: map[string]ds.IndexField{"Email": {IndField: 1}},
				Conditions: map[int]ds.IndexCondition{
					2: {ConditionType: "=", Value: []string{"active"}},
					3: {ConditionType: "=", Value: []string{"1"}, FieldExpression: "Flags&1"},
				},
			},
		}

		return pkg
	}

	first, err := g.GenerateFullSchema(newPkg())
	if err != nil {
		t.Fatalf("GenerateFullSchema error: %v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := g.GenerateFullSchema(newPkg())
		if err != nil {
			t.Fatalf("GenerateFullSchema error: %v", err)
		}

		if string(got) != string(first) {
			t.Fatalf("iteration %d: DDL отличается от первого прогона:\n%s\n---\n%s", i, first, got)
		}
	}

	if !strings.Contains(string(first), "WHERE status = 'active' AND flags&1 = 1") {
		t.Errorf("ожидался WHERE в порядке номеров полей, получено:\n%s", first)
	}
}
