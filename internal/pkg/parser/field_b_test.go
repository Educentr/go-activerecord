package parser

import (
	"go/ast"
	"reflect"
	"testing"

	"github.com/mailru/activerecord/internal/pkg/ds"
	"github.com/mailru/activerecord/pkg/activerecord"
)

func TestParseFields(t *testing.T) {
	type args struct {
		fields []*ast.Field
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    ds.RecordPackage
	}{
		{
			name: "simple fields",
			args: args{
				fields: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "ID"}},
						Type:  &ast.Ident{Name: "int"},
						Tag:   &ast.BasicLit{Value: "`" + `ar:"primary_key"` + "`"},
					},
					{
						Names: []*ast.Ident{{Name: "BarID"}},
						Type:  &ast.Ident{Name: "int"},
						Tag:   &ast.BasicLit{Value: "`" + `ar:""` + "`"},
					},
				},
			},
			wantErr: false,
			want: ds.RecordPackage{
				Server:    ds.ServerDeclaration{},
				Namespace: ds.NamespaceDeclaration{},
				Fields: []ds.FieldDeclaration{
					{Name: "ID", Format: "int", PrimaryKey: true, Mutators: []string{}, Serializer: []string{}},
					{Name: "BarID", Format: "int", PrimaryKey: false, Mutators: []string{}, Serializer: []string{}},
				},
				FieldsMap:       map[string]int{"ID": 0, "BarID": 1},
				FieldsObjectMap: map[string]ds.FieldObject{},
				Indexes: []ds.IndexDeclaration{
					{
						Name:     "ID",
						Num:      0,
						Selector: "SelectByID",
						Fields:   []int{0},
						FieldsMap: map[string]ds.IndexField{
							"ID": {IndField: 0, Order: 0},
						},
						Primary: true,
						Unique:  true,
					},
				},
				IndexMap:      map[string]int{"ID": 0},
				SelectorMap:   map[string]int{"SelectByID": 0},
				ImportPackage: ds.NewImportPackage(),
				Backends:      []activerecord.Backend{},
				SerializerMap: map[string]ds.SerializerDeclaration{},
				TriggerMap:    map[string]ds.TriggerDeclaration{},
				FlagMap:       map[string]ds.FlagDeclaration{},
			},
		},
	}

	rp := ds.NewRecordPackage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseFields(rp, tt.args.fields); (err != nil) != tt.wantErr {
				t.Errorf("ParseFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(rp.Indexes, tt.want.Indexes) {
				t.Errorf("ParseFields() = %+v, want %+v", rp, tt.want)
			}
		})
	}
}
