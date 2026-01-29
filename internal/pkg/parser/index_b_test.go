package parser_test

import (
	"go/ast"
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/parser"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestParseIndex(t *testing.T) {
	type args struct {
		dst    *ds.RecordPackage
		fields []*ast.Field
	}

	wantRp := ds.NewRecordPackage()
	wantRp.Fields = []ds.FieldDeclaration{
		{Name: "WeekNum", Format: "int"},
		{Name: "Amount", Format: "int"},
	}
	wantRp.FieldsMap = map[string]int{"WeekNum": 0, "Amount": 1}
	wantRp.Indexes = []ds.IndexDeclaration{
		{
			Name:     "WeekAmount",
			Num:      0,
			Selector: "SelectByWeekAmount",
			Fields:   []int{0, 1},
			FieldsMap: map[string]ds.IndexField{
				"WeekNum": {IndField: 0, Order: 0},
				"Amount":  {IndField: 1, Order: 1},
			},
			Unique:     false,
			Conditions: map[int]ds.IndexCondition{},
		},
	}
	wantRp.IndexMap = map[string]int{"WeekAmount": 0}
	wantRp.SelectorMap = map[string]int{"SelectByWeekAmount": 0}

	rp := ds.NewRecordPackage()

	err := rp.AddField(ds.FieldDeclaration{
		Name:       "WeekNum",
		Format:     "int",
		PrimaryKey: false,
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	err = rp.AddField(ds.FieldDeclaration{
		Name:       "Amount",
		Format:     "int",
		PrimaryKey: false,
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *ds.RecordPackage
	}{
		{
			name: "simple index",
			args: args{
				dst: rp,
				fields: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "WeekAmount"}},
						Type:  &ast.Ident{Name: "bool"},
						Tag:   &ast.BasicLit{Value: "`" + `ar:"fields:WeekNum,Amount=desc"` + "`"},
					},
				},
			},
			wantErr: false,
			want:    wantRp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parser.ParseIndexes(tt.args.dst, tt.args.fields); (err != nil) != tt.wantErr {
				t.Errorf("ParseIndexPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Check(t, cmp.DeepEqual(tt.want, tt.args.dst), "Invalid response, test `%s`", tt.name)
		})
	}
}

func TestParseIndexWithConditions(t *testing.T) {
	// Подготовка полей
	baseRp := ds.NewRecordPackage()
	_ = baseRp.AddField(ds.FieldDeclaration{Name: "Status", Format: "string"})
	_ = baseRp.AddField(ds.FieldDeclaration{Name: "Type", Format: "string"})
	_ = baseRp.AddField(ds.FieldDeclaration{Name: "DeletedAt", Format: "*time.Time"})
	_ = baseRp.AddField(ds.FieldDeclaration{Name: "Flags", Format: "int"})

	tests := []struct {
		name      string
		tag       string
		wantConds map[int]ds.IndexCondition
		wantErr   bool
	}{
		{
			name: "single condition",
			tag:  "`" + `ar:"fields:Status;condition:Status[=]'active'"` + "`",
			wantConds: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"active"}, IsNullCheck: false},
			},
		},
		{
			name: "two conditions with &&",
			tag:  "`" + `ar:"fields:Status,Type;condition:Status[=]''&&Type[!=]''"` + "`",
			wantConds: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{""}, IsNullCheck: false},
				1: {ConditionType: "!=", Value: []string{""}, IsNullCheck: false},
			},
		},
		{
			name: "three conditions",
			tag:  "`" + `ar:"fields:Status;condition:Status[=]1&&Type[is not null]&&DeletedAt[is null]"` + "`",
			wantConds: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"1"}, IsNullCheck: false},
				1: {ConditionType: "is not null", Value: []string{}, IsNullCheck: true},
				2: {ConditionType: "is null", Value: []string{}, IsNullCheck: true},
			},
		},
		{
			name: "condition with multiple values (IN clause)",
			tag:  "`" + `ar:"fields:Status;condition:Status[=]'active','pending','done'"` + "`",
			wantConds: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"active", "pending", "done"}, IsNullCheck: false},
			},
		},
		{
			name: "multiple conditions with multiple values",
			tag:  "`" + `ar:"fields:Status,Type;condition:Status[=]'a','b'&&Type[!=]'x'"` + "`",
			wantConds: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"a", "b"}, IsNullCheck: false},
				1: {ConditionType: "!=", Value: []string{"x"}, IsNullCheck: false},
			},
		},
		{
			name: "bitwise condition with &&",
			tag:  "`" + `ar:"fields:Flags;condition:Flags&1[=]1&&DeletedAt[is null]"` + "`",
			wantConds: map[int]ds.IndexCondition{
				3: {ConditionType: "=", Value: []string{"1"}, IsNullCheck: false, FieldExpression: "Flags&1"},
				2: {ConditionType: "is null", Value: []string{}, IsNullCheck: true},
			},
		},
		{
			name: "condition with other tags",
			tag:  "`" + `ar:"fields:Status,Type;unique;condition:Status[=]1&&Type[!=]''"` + "`",
			wantConds: map[int]ds.IndexCondition{
				0: {ConditionType: "=", Value: []string{"1"}, IsNullCheck: false},
				1: {ConditionType: "!=", Value: []string{""}, IsNullCheck: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRp := ds.NewRecordPackage()
			// Копируем поля
			for _, f := range baseRp.Fields {
				_ = testRp.AddField(f)
			}

			fields := []*ast.Field{{
				Names: []*ast.Ident{{Name: "TestIndex"}},
				Type:  &ast.Ident{Name: "bool"},
				Tag:   &ast.BasicLit{Value: tt.tag},
			}}

			err := parser.ParseIndexes(testRp, fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIndexes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(testRp.Indexes) > 0 {
				assert.Check(t, cmp.DeepEqual(tt.wantConds, testRp.Indexes[0].Conditions))
			}
		})
	}
}

func TestParseIndexPartWithConditions(t *testing.T) {
	rp := ds.NewRecordPackage()
	_ = rp.AddField(ds.FieldDeclaration{Name: "Status", Format: "string"})
	_ = rp.AddField(ds.FieldDeclaration{Name: "CreatedAt", Format: "time.Time"})
	_ = rp.AddField(ds.FieldDeclaration{Name: "Error", Format: "string"})

	conditions := map[int]ds.IndexCondition{
		0: {ConditionType: "=", Value: []string{"active"}, IsNullCheck: false},
		2: {ConditionType: "is null", Value: []string{}, IsNullCheck: true},
	}

	_ = rp.AddIndex(ds.IndexDeclaration{
		Name:     "StatusCreated",
		Num:      0,
		Selector: "SelectByStatusCreated",
		Fields:   []int{0, 1},
		FieldsMap: map[string]ds.IndexField{
			"Status":    {IndField: 0, Order: ds.IndexOrderAsc},
			"CreatedAt": {IndField: 1, Order: ds.IndexOrderAsc},
		},
		Conditions: conditions,
	})

	fields := []*ast.Field{
		{
			Names: []*ast.Ident{{Name: "StatusPart"}},
			Type:  &ast.Ident{Name: "bool"},
			Tag:   &ast.BasicLit{Value: "`" + `ar:"index:StatusCreated;fieldnum:1;selector:SelectByStatus"` + "`"},
		},
	}

	err := parser.ParseIndexPart(rp, fields)
	if err != nil {
		t.Fatalf("ParseIndexPart() unexpected error: %v", err)
	}

	if len(rp.Indexes) != 2 {
		t.Fatalf("expected 2 indexes, got %d", len(rp.Indexes))
	}

	partIndex := rp.Indexes[1]
	assert.Check(t, cmp.DeepEqual(conditions, partIndex.Conditions), "IndexPart should inherit Conditions from parent index")
	assert.Check(t, partIndex.Partial, "IndexPart should have Partial=true")
	assert.Check(t, cmp.Equal(1, len(partIndex.Fields)), "IndexPart should have 1 field")
}

func TestParseIndexPart(t *testing.T) {
	type args struct {
		dst    *ds.RecordPackage
		fields []*ast.Field
	}

	wantRp := ds.NewRecordPackage()
	wantRp.Fields = []ds.FieldDeclaration{
		{Name: "Field1", Format: "int"},
		{Name: "Field2", Format: "int"},
	}
	wantRp.FieldsMap = map[string]int{"Field1": 0, "Field2": 1}
	wantRp.Indexes = []ds.IndexDeclaration{
		{
			Name:     "Field1Field2",
			Num:      0,
			Selector: "SelectByField1Field2",
			Fields:   []int{0, 1},
			FieldsMap: map[string]ds.IndexField{
				"Field1": {IndField: 0, Order: 0},
				"Field2": {IndField: 1, Order: 0},
			},
			Unique: true,
		},
		{
			Name:     "Field1Part",
			Num:      0,
			Selector: "SelectByField1",
			Fields:   []int{0},
			FieldsMap: map[string]ds.IndexField{
				"Field1": {IndField: 0, Order: 0},
			},
			Partial: true,
		},
	}
	wantRp.IndexMap = map[string]int{"Field1Field2": 0, "Field1Part": 1}
	wantRp.SelectorMap = map[string]int{"SelectByField1": 1, "SelectByField1Field2": 0}

	rp := ds.NewRecordPackage()

	err := rp.AddField(ds.FieldDeclaration{
		Name:       "Field1",
		Format:     "int",
		PrimaryKey: false,
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	err = rp.AddField(ds.FieldDeclaration{
		Name:       "Field2",
		Format:     "int",
		PrimaryKey: false,
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	err = rp.AddIndex(ds.IndexDeclaration{
		Name:      "Field1Field2",
		Num:       0,
		Selector:  "SelectByField1Field2",
		Fields:    []int{0, 1},
		FieldsMap: map[string]ds.IndexField{"Field1": {IndField: 0, Order: ds.IndexOrderAsc}, "Field2": {IndField: 1, Order: ds.IndexOrderAsc}},
		Primary:   false,
		Unique:    true,
		Type:      "",
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *ds.RecordPackage
	}{
		{
			name: "simple index part",
			args: args{
				dst: rp,
				fields: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "Field1Part"}},
						Type:  &ast.Ident{Name: "bool"},
						Tag:   &ast.BasicLit{Value: "`" + `ar:"index:Field1Field2;fieldnum:1;selector:SelectByField1"` + "`"},
					},
				},
			},
			wantErr: false,
			want:    wantRp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parser.ParseIndexPart(tt.args.dst, tt.args.fields); (err != nil) != tt.wantErr {
				t.Errorf("ParseIndexPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Check(t, cmp.DeepEqual(tt.want, tt.args.dst), "Invalid response, test `%s`", tt.name)
		})
	}
}
