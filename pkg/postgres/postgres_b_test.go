package postgres_test

import (
	"reflect"
	"testing"

	"github.com/mailru/activerecord/pkg/postgres"
)

func TestGenerateSelect(t *testing.T) {
	type args struct {
		tableName  string
		fieldNames []string
		index      postgres.Index
		keys       [][]any
		offset     uint16
		limit      uint16
		cursor     postgres.CursorPosition
	}
	tests := []struct {
		name    string
		args    args
		want    *postgres.Query
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "email"},
				index: postgres.Index{
					Unique: true,
					Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
				},
				keys:   [][]any{{1}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString: `SELECT id, name, email FROM "users" WHERE id = $1`,
				Params:      []any{1},
			},
			wantErr: false,
		},
		{
			name: "bulk",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "email"},
				index: postgres.Index{
					Unique: true,
					Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
				},
				keys:   [][]any{{1}, {2}, {3}},
				offset: 0,
				limit:  10,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString: `SELECT id, name, email FROM "users" WHERE id IN ($1, $2, $3) ORDER BY id ASC LIMIT 10`,
				Params:      []any{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "bulk_multi_field",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "parent", "name", "email"},
				index: postgres.Index{
					Unique: true,
					Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}, postgres.OrderField{Field: "parent", Order: postgres.DESC}},
				},
				keys:   [][]any{{1, 1}, {2, 1}, {3, 4}},
				offset: 2,
				limit:  5,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString: `SELECT id, parent, name, email FROM "users" WHERE (id, parent) IN (($1, $2), ($3, $4), ($5, $6)) ORDER BY id ASC, parent DESC LIMIT 5 OFFSET 2`,
				Params:      []any{1, 1, 2, 1, 3, 4},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := postgres.GenerateSelect(tt.args.tableName, tt.args.fieldNames, tt.args.index, tt.args.keys, tt.args.offset, tt.args.limit, tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSelect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}
