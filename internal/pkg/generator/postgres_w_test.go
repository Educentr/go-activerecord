package generator

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/backend"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

const (
	// Права на каталог и файлы временного пакета, который собирается в тесте
	testPkgDirPerm  = 0o750
	testPkgFilePerm = 0o600
)

// conditionPkgData описывает пакет с частичными индексами по полям разных
// типов: значение условия для каждого типа должно попасть в index.go литералом
// этого типа
func conditionPkgData() PkgData {
	fields := []ds.FieldDeclaration{
		{Name: "ID", Format: "int64", PrimaryKey: true, Mutators: []string{}, Serializer: []string{}},
		{Name: "VmID", Format: "int64", Mutators: []string{}, Serializer: []string{}},
		{Name: "Status", Format: "string", Size: 32, Mutators: []string{}, Serializer: []string{}},
		{Name: "Ratio", Format: "float64", Mutators: []string{}, Serializer: []string{}},
		{Name: "Enabled", Format: "bool", Mutators: []string{}, Serializer: []string{}},
		{Name: "CreatedAt", Format: "time.Time", Mutators: []string{}, Serializer: []string{}},
		{Name: "Raw", Format: "postgres.BYTEA", Mutators: []string{}, Serializer: []string{}},
		{Name: "Token", Format: "uuid.UUID", Mutators: []string{}, Serializer: []string{}},
	}

	fieldsMap := make(map[string]int, len(fields))
	for num, fld := range fields {
		fieldsMap[fld.Name] = num
	}

	indexes := []ds.IndexDeclaration{
		{
			Name: "ID", Num: 0, Selector: "SelectByID", Fields: []int{0},
			FieldsMap: map[string]ds.IndexField{"ID": {IndField: 0, Order: ds.IndexOrderAsc}},
			Primary:   true, Unique: true, Type: "int64",
		},
		{
			Name: "VmID", Num: 1, Selector: "SelectByVmID", Fields: []int{1},
			FieldsMap: map[string]ds.IndexField{"VmID": {IndField: 1, Order: ds.IndexOrderAsc}},
			Unique:    true, Type: "int64", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				1: {ConditionType: "!=", Value: []string{"0"}},
			},
		},
		{
			Name: "Status", Num: 2, Selector: "SelectByStatus", Fields: []int{2},
			FieldsMap: map[string]ds.IndexField{"Status": {IndField: 2, Order: ds.IndexOrderAsc}},
			Type:      "string", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				2: {ConditionType: "=", Value: []string{"active"}},
			},
		},
		{
			Name: "Ratio", Num: 3, Selector: "SelectByRatio", Fields: []int{3},
			FieldsMap: map[string]ds.IndexField{"Ratio": {IndField: 3, Order: ds.IndexOrderAsc}},
			Type:      "float64", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				3: {ConditionType: ">", Value: []string{"0.5"}},
			},
		},
		{
			Name: "Enabled", Num: 4, Selector: "SelectByEnabled", Fields: []int{4},
			FieldsMap: map[string]ds.IndexField{"Enabled": {IndField: 4, Order: ds.IndexOrderAsc}},
			Type:      "bool", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				4: {ConditionType: "=", Value: []string{"true"}},
			},
		},
		{
			Name: "CreatedAt", Num: 5, Selector: "SelectByCreatedAt", Fields: []int{5},
			FieldsMap: map[string]ds.IndexField{"CreatedAt": {IndField: 5, Order: ds.IndexOrderAsc}},
			Type:      "time.Time", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				5: {ConditionType: ">", Value: []string{"1700000000"}},
			},
		},
		{
			Name: "Raw", Num: 6, Selector: "SelectByRaw", Fields: []int{6},
			FieldsMap: map[string]ds.IndexField{"Raw": {IndField: 6, Order: ds.IndexOrderAsc}},
			Type:      "postgres.BYTEA", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				6: {ConditionType: "!=", Value: []string{`\x00`}},
			},
		},
		{
			Name: "Token", Num: 7, Selector: "SelectByToken", Fields: []int{7},
			FieldsMap: map[string]ds.IndexField{"Token": {IndField: 7, Order: ds.IndexOrderAsc}},
			Unique:    true, Type: "uuid.UUID", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				7: {ConditionType: "!=", Value: []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}},
			},
		},
		{
			Name: "Missing", Num: 8, Selector: "SelectByMissing", Fields: []int{2},
			FieldsMap: map[string]ds.IndexField{"Status": {IndField: 2, Order: ds.IndexOrderAsc}},
			Type:      "string", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				2: {ConditionType: "is not null", IsNullCheck: true, Value: []string{}},
			},
		},
	}

	return PkgData{
		ARPkg:            "providerkey",
		ARPkgTitle:       "ProviderKey",
		FieldList:        fields,
		FieldMap:         fieldsMap,
		FieldObject:      map[string]ds.FieldObject{},
		LinkedObject:     map[string]ds.RecordPackage{},
		ProcInFieldList:  []ds.ProcFieldDeclaration{},
		ProcOutFieldList: []ds.ProcFieldDeclaration{},
		ServerConfKey:    "testGenerateConfKey",
		Container:        ds.NamespaceDeclaration{ObjectName: "provider_keys", PublicName: "ProviderKey", PackageName: "providerkey"},
		Indexes:          indexes,
		Serializers:      map[string]ds.SerializerDeclaration{},
		Mutators:         map[string]ds.MutatorDeclaration{},
		Imports: []ds.ImportDeclaration{
			{Path: "github.com/gofrs/uuid/v5", ImportName: "uuid"},
		},
		Triggers: map[string]ds.TriggerDeclaration{},
		Flags:    map[string]ds.FlagDeclaration{},
		AppInfo:  "test",
	}
}

func generatePostgresPkg(t *testing.T, params PkgData) map[string][]byte {
	t.Helper()

	backend.RegisterBackend()

	backendObj, errB := backend.GetBackendByName("postgres")
	if errB != nil {
		t.Fatalf("can't get postgres backend: %s", errB)
	}

	ret, err := GenerateFromDir(params, backendObj)
	if err != nil {
		t.Fatalf("GenerateFromDir() error = %s", err)
	}

	return ret
}

// TestGeneratePostgresIndexCondition проверяет, что значение условия частичного
// индекса попадает в index.go литералом типа поля. Раньше оно подставлялось
// вызовом парсера строки (например int64(strconv.ParseInt("0", 10, 64))), и
// сгенерированный пакет не компилировался: такой вызов возвращает два значения,
// а в элементе []any допустимо только одно.
func TestGeneratePostgresIndexCondition(t *testing.T) {
	generated := generatePostgresPkg(t, conditionPkgData())

	index, ex := generated["index"]
	if !ex {
		t.Fatal("index.go not generated")
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "index.go", index, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("generated index.go is not valid go: %s\n%s", err, index)
	}

	want := map[string][]string{
		"indexID":        nil,
		"indexVmID":      {"int64(0)"},
		"indexStatus":    {`"active"`},
		"indexRatio":     {"float64(0.5)"},
		"indexEnabled":   {"true"},
		"indexCreatedAt": {"time.Unix(1700000000, 0).UTC()"},
		"indexRaw":       {`postgres.BYTEA("\\x00")`},
		"indexToken":     {"uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}"},
		"indexMissing":   nil,
	}

	for name, wantValues := range want {
		got, ex := conditionValues(t, fset, file, name)
		if !ex {
			t.Errorf("index var `%s` not found in generated index.go", name)
			continue
		}

		if strings.Join(got, ", ") != strings.Join(wantValues, ", ") {
			t.Errorf("`%s` condition values = %v, want %v", name, got, wantValues)
		}
	}
}

// TestGeneratePostgresCompiles собирает сгенерированный пакет компилятором:
// ошибки типов (а именно ими был баг с условием индекса) видны только так,
// синтаксический разбор их не ловит
func TestGeneratePostgresCompiles(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("go toolchain not available: %s", err)
	}

	generated := generatePostgresPkg(t, conditionPkgData())

	// Пакет собирается внутри модуля, иначе не разрешаются импорты
	// самого go-activerecord
	if err := os.MkdirAll("testdata", testPkgDirPerm); err != nil {
		t.Fatalf("can't create testdata dir: %s", err)
	}

	pkgDir, err := os.MkdirTemp("testdata", "compile")
	if err != nil {
		t.Fatalf("can't create package dir: %s", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(pkgDir); err != nil {
			t.Logf("can't remove %s: %s", pkgDir, err)
		}
	})

	for name, data := range generated {
		if err := os.WriteFile(filepath.Join(pkgDir, name+".go"), data, testPkgFilePerm); err != nil {
			t.Fatalf("can't write %s: %s", name, err)
		}
	}

	cmd := exec.Command(goBin, "build", "./"+filepath.ToSlash(pkgDir))
	cmd.Env = os.Environ()

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("generated package does not compile: %s\n%s", err, out)
	}
}

// conditionValues вытаскивает исходный текст значений условия у переменной
// индекса: сравнивать надо именно литералы, как они попали в файл
func conditionValues(t *testing.T, fset *token.FileSet, file *ast.File, varName string) ([]string, bool) {
	t.Helper()

	lit, ex := indexLiteral(file, varName)
	if !ex {
		return nil, false
	}

	values := []string{}

	for _, cond := range conditionLiterals(lit) {
		for _, kv := range cond.Elts {
			field, ok := kv.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			if key, ok := field.Key.(*ast.Ident); !ok || key.Name != "Values" {
				continue
			}

			valuesLit, ok := field.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}

			for _, value := range valuesLit.Elts {
				values = append(values, exprString(t, fset, value))
			}
		}
	}

	return values, true
}

// indexLiteral находит композитный литерал postgres.Index у переменной индекса
func indexLiteral(file *ast.File, varName string) (*ast.CompositeLit, bool) {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != varName {
				continue
			}

			if len(valueSpec.Values) != 1 {
				continue
			}

			if lit, ok := valueSpec.Values[0].(*ast.CompositeLit); ok {
				return lit, true
			}
		}
	}

	return nil, false
}

// conditionLiterals возвращает литералы postgres.Condition из поля Condition
func conditionLiterals(index *ast.CompositeLit) []*ast.CompositeLit {
	ret := []*ast.CompositeLit{}

	for _, elt := range index.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		if key, ok := kv.Key.(*ast.Ident); !ok || key.Name != "Condition" {
			continue
		}

		conditions, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, cond := range conditions.Elts {
			if lit, ok := cond.(*ast.CompositeLit); ok {
				ret = append(ret, lit)
			}
		}
	}

	return ret
}

func exprString(t *testing.T, fset *token.FileSet, expr ast.Expr) string {
	t.Helper()

	buf := &strings.Builder{}
	if err := printer.Fprint(buf, fset, expr); err != nil {
		t.Fatalf("can't print expression: %s", err)
	}

	return buf.String()
}
