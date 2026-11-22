package postgres

import (
	"strings"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

// conditionRecordPackage собирает декларацию с одним частичным индексом по полю
// VmID типа int64
func conditionRecordPackage(conditionType string, values []string) *ds.RecordPackage {
	rp := ds.NewRecordPackage()

	rp.Namespace = ds.NamespaceDeclaration{ObjectName: "provider_keys", PublicName: "ProviderKey", PackageName: "providerkey"}
	rp.Fields = []ds.FieldDeclaration{
		{Name: "ID", Format: "int64", PrimaryKey: true, Mutators: []string{}, Serializer: []string{}},
		{Name: "VmID", Format: "int64", Mutators: []string{}, Serializer: []string{}},
	}
	rp.FieldsMap = map[string]int{"ID": 0, "VmID": 1}
	rp.Indexes = []ds.IndexDeclaration{
		{
			Name: "VmID", Num: 0, Selector: "SelectByVmID", Fields: []int{1},
			FieldsMap: map[string]ds.IndexField{"VmID": {IndField: 1, Order: ds.IndexOrderAsc}},
			Unique:    true, Type: "int64", Partial: true,
			Conditions: map[int]ds.IndexCondition{
				1: {ConditionType: conditionType, Value: values},
			},
		},
	}

	return rp
}

// TestCheckIndexesConditionValue проверяет, что константа условия, которую
// нельзя привести к типу поля, отбраковывается на проверке декларации: в
// сгенерированный код она попадает литералом, и до компиляции такое значение
// доходить не должно
func TestCheckIndexesConditionValue(t *testing.T) {
	tests := []struct {
		name          string
		conditionType string
		values        []string
		wantErr       bool
	}{
		{name: "valid int", conditionType: "!=", values: []string{"0"}},
		{name: "valid negative int", conditionType: ">", values: []string{"-1"}},
		{name: "not a number", conditionType: "!=", values: []string{"abc"}, wantErr: true},
		{name: "float in int field", conditionType: ">", values: []string{"0.5"}, wantErr: true},
		{name: "second value invalid", conditionType: "!=", values: []string{"0", "abc"}, wantErr: true},
	}

	backendGen := BackendGenerator{}
	backendGen.Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := backendGen.CheckIndexes(conditionRecordPackage(tt.conditionType, tt.values))
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckIndexes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				return
			}

			// В сообщении должно быть видно, какое поле и какое значение чинить
			if !strings.Contains(err.Error(), "VmID") || !strings.Contains(err.Error(), "int64") {
				t.Errorf("CheckIndexes() error = %q, want it to mention field and type", err)
			}
		})
	}
}
