package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/testutil"
)

const schemaRunsCount = 10

// addDeterminismField добавляет поле в пакет, прерывая тест при ошибке
func addDeterminismField(t *testing.T, pkg *ds.RecordPackage, field ds.FieldDeclaration) {
	t.Helper()

	field.Mutators = []string{}
	field.Serializer = []string{}

	if err := pkg.AddField(field); err != nil {
		t.Fatalf("AddField(%s) error: %v", field.Name, err)
	}
}

// newDeterminismPackages собирает три postgres-таблицы: порядок их объявления в
// map не совпадает с алфавитным, у одной таблицы — условный индекс с двумя
// условиями и внешние ключи
func newDeterminismPackages(t *testing.T) map[string]*ds.RecordPackage {
	t.Helper()

	users := ds.NewRecordPackage()
	users.Backends = []ds.Backend{"postgres"}
	users.ServerConfKey = "testconf"
	users.Namespace = ds.NamespaceDeclaration{ObjectName: "users", PackageName: "users", PublicName: "Users"}
	addDeterminismField(t, users, ds.FieldDeclaration{Name: "Id", Format: "int64", PrimaryKey: true})
	addDeterminismField(t, users, ds.FieldDeclaration{Name: "Email", Format: "string", Size: 255})
	addDeterminismField(t, users, ds.FieldDeclaration{Name: "Status", Format: "string", Size: 32})
	addDeterminismField(t, users, ds.FieldDeclaration{Name: "Flags", Format: "uint32"})
	users.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
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

	products := ds.NewRecordPackage()
	products.Backends = []ds.Backend{"postgres"}
	products.ServerConfKey = "testconf"
	products.Namespace = ds.NamespaceDeclaration{ObjectName: "products", PackageName: "products", PublicName: "Products"}
	addDeterminismField(t, products, ds.FieldDeclaration{Name: "Id", Format: "int64", PrimaryKey: true})
	addDeterminismField(t, products, ds.FieldDeclaration{Name: "Name", Format: "string", Size: 100})
	products.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	orders := ds.NewRecordPackage()
	orders.Backends = []ds.Backend{"postgres"}
	orders.ServerConfKey = "testconf"
	orders.Namespace = ds.NamespaceDeclaration{ObjectName: "orders", PackageName: "orders", PublicName: "Orders"}
	addDeterminismField(t, orders, ds.FieldDeclaration{Name: "Id", Format: "int64", PrimaryKey: true})
	addDeterminismField(t, orders, ds.FieldDeclaration{Name: "UserId", Format: "int64"})
	addDeterminismField(t, orders, ds.FieldDeclaration{Name: "ProductId", Format: "int64"})
	orders.FieldsObjectMap = map[string]ds.FieldObject{
		"User":    {Name: "User", Key: "Id", ObjectName: "users", Field: "UserId"},
		"Product": {Name: "Product", Key: "Id", ObjectName: "products", Field: "ProductId"},
	}
	orders.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0},
			FieldsMap: map[string]ds.IndexField{"Id": {IndField: 0}},
			Primary:   true,
			Unique:    true,
		},
	}

	return map[string]*ds.RecordPackage{
		"users":    users,
		"products": products,
		"orders":   orders,
	}
}

// Повторный прогон генерации схемы на неизменённых декларациях даёт побайтно
// одинаковые schema.sql и schema.json и не создаёт новых файлов миграций
func TestArGen_generateSchemaArtifacts_Deterministic(t *testing.T) {
	tempDirs := testutil.InitTmps()
	defer tempDirs.Defer()

	src, dst, err := tempDirs.CreateDirs(testutil.EmptyDstDir)
	if err != nil {
		t.Fatalf("CreateDirs error: %v", err)
	}

	schemaDir, err := tempDirs.AddTempDir()
	if err != nil {
		t.Fatalf("AddTempDir error: %v", err)
	}

	argen, err := Init(context.Background(), &testutil.TestAppInfo, src, dst, "", "github.com/Educentr/go-activerecord", Options{
		SchemaPath: schemaDir,
	})
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	pgSchemaDir := filepath.Join(schemaDir, "postgres", "testconf")
	schemaSQLPath := filepath.Join(pgSchemaDir, "schema.sql")
	schemaJSONPath := filepath.Join(pgSchemaDir, "schema.json")
	migrationsDir := filepath.Join(pgSchemaDir, "migrations")

	var firstSQL, firstJSON []byte

	for run := 1; run <= schemaRunsCount; run++ {
		// Декларации пересобираются на каждом прогоне: так же, как при новом
		// запуске argen, где порядок обхода map снова случайный
		argen.packagesParsed = newDeterminismPackages(t)

		if err := argen.generateSchemaArtifacts(); err != nil {
			t.Fatalf("run %d: generateSchemaArtifacts error: %v", run, err)
		}

		gotSQL, err := os.ReadFile(schemaSQLPath)
		if err != nil {
			t.Fatalf("run %d: ReadFile schema.sql error: %v", run, err)
		}

		gotJSON, err := os.ReadFile(schemaJSONPath)
		if err != nil {
			t.Fatalf("run %d: ReadFile schema.json error: %v", run, err)
		}

		if run == 1 {
			firstSQL, firstJSON = gotSQL, gotJSON

			continue
		}

		if string(gotSQL) != string(firstSQL) {
			t.Fatalf("run %d: schema.sql отличается от первого прогона:\n%s\n---\n%s", run, firstSQL, gotSQL)
		}

		if string(gotJSON) != string(firstJSON) {
			t.Fatalf("run %d: schema.json отличается от первого прогона:\n%s\n---\n%s", run, firstJSON, gotJSON)
		}

		// Схема не менялась, значит и новых миграций быть не должно
		entries, err := os.ReadDir(migrationsDir)
		if err != nil && !os.IsNotExist(err) {
			t.Fatalf("run %d: ReadDir migrations error: %v", run, err)
		}

		if len(entries) != 0 {
			names := make([]string, 0, len(entries))
			for _, entry := range entries {
				names = append(names, entry.Name())
			}

			t.Fatalf("run %d: на неизменённых декларациях созданы миграции: %v", run, names)
		}
	}

	// Блоки таблиц идут в алфавитном порядке имён таблиц
	sql := string(firstSQL)
	wantOrder := []string{"orders", "products", "users"}
	prev := -1

	for _, tableName := range wantOrder {
		idx := strings.Index(sql, "CREATE TABLE IF NOT EXISTS "+tableName+" (")
		if idx < 0 {
			t.Fatalf("в schema.sql нет блока таблицы %s:\n%s", tableName, sql)
		}

		if idx < prev {
			t.Fatalf("блоки таблиц в schema.sql идут не по алфавиту (%s раньше ожидаемого места):\n%s", tableName, sql)
		}

		prev = idx
	}

	// Условия индекса — в порядке номеров полей
	if !strings.Contains(sql, "WHERE status = 'active' AND flags&1 = 1") {
		t.Errorf("ожидался WHERE в порядке номеров полей, получено:\n%s", sql)
	}

	// Внешние ключи — в порядке имён полей связи
	productFK := strings.Index(sql, "CONSTRAINT orders_productid_fkey")
	userFK := strings.Index(sql, "CONSTRAINT orders_userid_fkey")

	if productFK < 0 || userFK < 0 {
		t.Fatalf("в schema.sql нет ожидаемых CONSTRAINT:\n%s", sql)
	}

	if productFK > userFK {
		t.Errorf("CONSTRAINT идут не в порядке имён полей связи:\n%s", sql)
	}
}
