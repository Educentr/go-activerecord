package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/schema"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/testutil"
)

// newOctopusSpacePackage собирает octopus-пакет с номером спейса из
// //ar:namespace:<spaceID> и ключом конфига сервера без номера в нём
func newOctopusSpacePackage(t *testing.T, packageName, publicName, spaceID string) *ds.RecordPackage {
	t.Helper()

	pkg := ds.NewRecordPackage()
	pkg.Backends = []ds.Backend{"octopus"}
	pkg.ServerConfKey = "ocdb"
	pkg.Namespace = ds.NamespaceDeclaration{ObjectName: spaceID, PackageName: packageName, PublicName: publicName}

	for _, field := range []ds.FieldDeclaration{
		{Name: "Field1", Format: "int", PrimaryKey: true, Mutators: []string{}, Serializer: []string{}},
		{Name: "Field2", Format: "string", PrimaryKey: true, Size: 255, Mutators: []string{}, Serializer: []string{}},
	} {
		if err := pkg.AddField(field); err != nil {
			t.Fatalf("AddField(%s) error: %v", field.Name, err)
		}
	}

	pkg.Indexes = []ds.IndexDeclaration{
		{
			Name:      "PK",
			Fields:    []int{0, 1},
			FieldsMap: map[string]ds.IndexField{"Field1": {IndField: 0}, "Field2": {IndField: 1}},
			Primary:   true,
			Unique:    true,
		},
	}

	return pkg
}

// newOctopusProcPackage собирает octopus-декларацию хранимой процедуры:
// полей нет, в //ar:namespace лежит имя процедуры, а не номер спейса
func newOctopusProcPackage(t *testing.T, packageName, procName string) *ds.RecordPackage {
	t.Helper()

	pkg := ds.NewRecordPackage()
	pkg.Backends = []ds.Backend{"octopus"}
	pkg.ServerConfKey = "ocdb"
	pkg.Namespace = ds.NamespaceDeclaration{ObjectName: procName, PackageName: packageName, PublicName: "Proc"}

	for _, field := range []ds.ProcFieldDeclaration{
		{Name: "Input", Format: "string", Type: ds.IN, OrderIndex: 0},
		{Name: "Output", Format: "string", Type: ds.OUT, OrderIndex: 0},
	} {
		if err := pkg.AddProcField(field); err != nil {
			t.Fatalf("AddProcField(%s) error: %v", field.Name, err)
		}
	}

	return pkg
}

// Номер спейса в space.cfg и schema.json берётся из //ar:namespace:<N>, а не из
// ServerConfKey: иначе все спейсы схлопывались в object_space[0] (LIB-5).
// Декларации процедур спейса не имеют и в схему не попадают
func TestArGen_generateSchemaArtifacts_OctopusSpaceIDFromNamespace(t *testing.T) {
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

	argen.packagesParsed = map[string]*ds.RecordPackage{
		"first":  newOctopusSpacePackage(t, "first", "First", "3"),
		"second": newOctopusSpacePackage(t, "second", "Second", "4"),
		"proc":   newOctopusProcPackage(t, "proc", "some_procedure"),
	}

	if err := argen.generateSchemaArtifacts(); err != nil {
		t.Fatalf("generateSchemaArtifacts error: %v", err)
	}

	octopusSchemaDir := filepath.Join(schemaDir, "octopus", "ocdb")

	cfg, err := os.ReadFile(filepath.Join(octopusSchemaDir, "space.cfg"))
	if err != nil {
		t.Fatalf("ReadFile space.cfg error: %v", err)
	}

	spaceCfg := string(cfg)

	for _, want := range []string{"# Space ID: 3", "object_space[3]", "# Space ID: 4", "object_space[4]"} {
		if !strings.Contains(spaceCfg, want) {
			t.Errorf("в space.cfg нет %q:\n%s", want, spaceCfg)
		}
	}

	// Пакеты-процедуры не дают блока спейса, в том числе нулевого
	if strings.Contains(spaceCfg, "object_space[0]") {
		t.Errorf("в space.cfg попал object_space[0]:\n%s", spaceCfg)
	}

	if got := strings.Count(spaceCfg, "object_space["); got != 2 {
		t.Errorf("ожидалось 2 блока object_space, получено %d:\n%s", got, spaceCfg)
	}

	// space_id в schema.json — тот же номер из //ar:namespace
	schemaJSON, err := os.ReadFile(filepath.Join(octopusSchemaDir, "schema.json"))
	if err != nil {
		t.Fatalf("ReadFile schema.json error: %v", err)
	}

	var saved schema.Schema
	if err := json.Unmarshal(schemaJSON, &saved); err != nil {
		t.Fatalf("Unmarshal schema.json error: %v", err)
	}

	wantSpaceIDs := map[string]uint{"First": 3, "Second": 4}

	if len(saved.Tables) != len(wantSpaceIDs) {
		t.Fatalf("ожидалось %d спейсов в schema.json, получено %d: %+v", len(wantSpaceIDs), len(saved.Tables), saved.Tables)
	}

	for name, wantID := range wantSpaceIDs {
		table, ex := saved.Tables[name]
		if !ex {
			t.Fatalf("в schema.json нет спейса %s: %+v", name, saved.Tables)
		}

		if table.SpaceID != wantID {
			t.Errorf("space_id спейса %s = %d, ожидался %d", name, table.SpaceID, wantID)
		}
	}
}
