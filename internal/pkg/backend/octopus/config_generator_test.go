package octopus

import (
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/schema"
)

func TestOctopusConfigGenerator_SupportsMigrations(t *testing.T) {
	g := NewConfigGenerator()
	if g.SupportsMigrations() {
		t.Error("Octopus should not support migrations")
	}
}

func TestOctopusConfigGenerator_SchemaFileName(t *testing.T) {
	g := NewConfigGenerator()
	if g.SchemaFileName() != "space.cfg" {
		t.Errorf("expected 'space.cfg', got '%s'", g.SchemaFileName())
	}
}

func TestOctopusConfigGenerator_SchemaFileExtension(t *testing.T) {
	g := NewConfigGenerator()
	if g.SchemaFileExtension() != ".cfg" {
		t.Errorf("expected '.cfg', got '%s'", g.SchemaFileExtension())
	}
}

func TestOctopusConfigGenerator_GenerateSchemaJSON(t *testing.T) {
	g := NewConfigGenerator()

	pkg := ds.NewRecordPackage()
	// ServerConfKey — ключ конфига сервера, номера спейса в нём нет:
	// спейс берётся из //ar:namespace:512 (Namespace.ObjectName)
	pkg.ServerConfKey = "ocdb"
	pkg.Namespace.ObjectName = "512"
	pkg.Namespace.PublicName = "users"
	pkg.Namespace.PackageName = "users"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Id", Format: "uint64", PrimaryKey: true},
		{Name: "Email", Format: "string", Size: 255},
		{Name: "Status", Format: "uint32"},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Num:       0,
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
		{
			Name:      "Email",
			Num:       1,
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
	if table.Backend != "octopus" {
		t.Errorf("expected backend 'octopus', got '%s'", table.Backend)
	}
	if table.SpaceID != 512 {
		t.Errorf("expected SpaceID 512, got %d", table.SpaceID)
	}
	if len(table.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(table.Columns))
	}
	if len(table.Indexes) != 2 {
		t.Errorf("expected 2 indexes, got %d", len(table.Indexes))
	}
}

func TestOctopusConfigGenerator_GenerateFullSchema(t *testing.T) {
	g := NewConfigGenerator()

	pkg := ds.NewRecordPackage()
	pkg.ServerConfKey = "ocdb"
	pkg.Namespace.ObjectName = "512"
	pkg.Namespace.PublicName = "users"
	pkg.Namespace.PackageName = "users"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Id", Format: "uint64", PrimaryKey: true},
		{Name: "Email", Format: "string", Size: 255},
	}
	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Num:       0,
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	cfg, err := g.GenerateFullSchema(pkg)
	if err != nil {
		t.Fatalf("GenerateFullSchema error: %v", err)
	}

	cfgStr := string(cfg)

	if !strings.Contains(cfgStr, "object_space[512]") {
		t.Error("expected config to contain object_space declaration")
	}

	if !strings.Contains(cfgStr, "# Space ID: 512") {
		t.Error("expected config comment to contain space ID from //ar:namespace")
	}
	if !strings.Contains(cfgStr, "enabled = 1") {
		t.Error("expected config to contain enabled = 1")
	}
	if !strings.Contains(cfgStr, "index[0]") {
		t.Error("expected config to contain index[0]")
	}
	if !strings.Contains(cfgStr, `type = "HASH"`) {
		t.Error("expected config to contain HASH type for primary key")
	}
	if !strings.Contains(cfgStr, "fieldno = 0") {
		t.Error("expected config to contain fieldno = 0")
	}
	if !strings.Contains(cfgStr, `type = "NUM64"`) {
		t.Error("expected config to contain NUM64 type for uint64 field")
	}
}

func TestOctopusConfigGenerator_GenerateMigration(t *testing.T) {
	g := NewConfigGenerator()

	diff := &schema.TableDiff{
		AddedColumns: []schema.Column{
			{Name: "NewField", Type: "NUM32"},
		},
	}

	migration, err := g.GenerateMigration(diff, "users")
	if err != nil {
		t.Fatalf("GenerateMigration error: %v", err)
	}

	if migration != nil {
		t.Error("expected nil migration for Octopus (does not support migrations)")
	}
}

func TestOctopusConfigGenerator_goTypeToOctopusType(t *testing.T) {
	g := NewConfigGenerator()

	tests := []struct {
		format ds.Format
		want   string
	}{
		{"uint64", "NUM64"},
		{"int64", "NUM64"},
		{"uint32", "NUM32"},
		{"int32", "NUM32"},
		{"uint16", "NUM16"},
		{"int16", "NUM16"},
		{"uint8", "NUM8"},
		{"int8", "NUM8"},
		{"string", "STR"},
		{"bool", "NUM8"},
	}

	for _, tt := range tests {
		got := g.goTypeToOctopusType(tt.format)
		if got != tt.want {
			t.Errorf("goTypeToOctopusType(%s) = %s, want %s", tt.format, got, tt.want)
		}
	}
}

// Номер спейса берётся из //ar:namespace и не зависит от ServerConfKey:
// ключ конфига сервера с числовым хвостом не должен подменять номер спейса
func TestOctopusConfigGenerator_SpaceIDFromNamespace(t *testing.T) {
	g := NewConfigGenerator()

	tests := []struct {
		name          string
		objectName    string
		serverConfKey string
		want          uint
	}{
		{name: "простой ключ конфига", objectName: "3", serverConfKey: "ocdb", want: 3},
		{name: "ключ конфига с числовым хвостом", objectName: "4", serverConfKey: "ocdb:512", want: 4},
		{name: "спейс 0 остаётся нулём", objectName: "0", serverConfKey: "ocdb:512", want: 0},
		{name: "пустой ключ конфига", objectName: "1024", serverConfKey: "", want: 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := ds.NewRecordPackage()
			pkg.ServerConfKey = tt.serverConfKey
			pkg.Namespace.ObjectName = tt.objectName
			pkg.Namespace.PublicName = "users"
			pkg.Namespace.PackageName = "users"
			pkg.Fields = []ds.FieldDeclaration{
				{Name: "Id", Format: "uint64", PrimaryKey: true},
			}

			table, err := g.GenerateSchemaJSON(pkg)
			if err != nil {
				t.Fatalf("GenerateSchemaJSON error: %v", err)
			}

			if table.SpaceID != tt.want {
				t.Errorf("SpaceID = %d, want %d", table.SpaceID, tt.want)
			}
		})
	}
}

// Нечисловой namespace — это не спейс: тихий 0 вместо ошибки дал бы
// невалидный object_space[0] в space.cfg
func TestOctopusConfigGenerator_GenerateSchemaJSON_NonNumericNamespace(t *testing.T) {
	g := NewConfigGenerator()

	pkg := ds.NewRecordPackage()
	pkg.ServerConfKey = "ocdb"
	pkg.Namespace.ObjectName = "users"
	pkg.Namespace.PublicName = "users"
	pkg.Namespace.PackageName = "users"
	pkg.Fields = []ds.FieldDeclaration{
		{Name: "Id", Format: "uint64", PrimaryKey: true},
	}

	table, err := g.GenerateSchemaJSON(pkg)
	if err == nil {
		t.Fatalf("expected error for non-numeric namespace, got table %+v", table)
	}

	if !strings.Contains(err.Error(), "users") {
		t.Errorf("expected error to mention namespace, got: %v", err)
	}
}
