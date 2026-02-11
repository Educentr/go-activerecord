package postgres_test

import (
	"reflect"
	"testing"

	"github.com/Educentr/go-activerecord/v3/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestOrderConditionsWithDir(t *testing.T) {
	tests := []struct {
		name  string
		index postgres.Index
		dir   postgres.Order
		want  string
	}{
		{
			name: "single field ASC + dir ASC",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			dir:  postgres.ASC,
			want: " ORDER BY id ASC",
		},
		{
			name: "single field ASC + dir DESC",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			dir:  postgres.DESC,
			want: " ORDER BY id DESC",
		},
		{
			name: "multi-field ASC,DESC + dir DESC reverses both",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
					postgres.OrderField{Field: "name", Order: postgres.DESC},
				},
			},
			dir:  postgres.DESC,
			want: " ORDER BY id DESC, name ASC",
		},
		{
			name: "multi-field ASC,DESC + dir ASC keeps original",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
					postgres.OrderField{Field: "name", Order: postgres.DESC},
				},
			},
			dir:  postgres.ASC,
			want: " ORDER BY id ASC, name DESC",
		},
		{
			name: "single field DESC + dir DESC keeps ASC",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "created_at", Order: postgres.DESC},
				},
			},
			dir:  postgres.DESC,
			want: " ORDER BY created_at ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.index.OrderConditionsWithDir(tt.dir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCursorConditionsWithOrder(t *testing.T) {
	tests := []struct {
		name         string
		index        postgres.Index
		cursor       postgres.CursorPosition
		paramsOffset int
		order        postgres.Order
		wantStr      string
		wantParams   []any
	}{
		{
			name: "single field ASC with values",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			cursor:       postgres.CursorPosition{Values: []any{42}},
			paramsOffset: 0,
			order:        postgres.ASC,
			wantStr:      " AND id > $1",
			wantParams:   []any{42},
		},
		{
			name: "single field DESC with values",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			cursor:       postgres.CursorPosition{Values: []any{42}},
			paramsOffset: 0,
			order:        postgres.DESC,
			wantStr:      " AND id < $1",
			wantParams:   []any{42},
		},
		{
			name: "multi-field DESC with values",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "col1", Order: postgres.ASC},
					postgres.OrderField{Field: "col2", Order: postgres.DESC},
				},
			},
			cursor:       postgres.CursorPosition{Values: []any{10, 20}},
			paramsOffset: 0,
			order:        postgres.DESC,
			wantStr:      " AND ( col1, col2) < ($1, $2)",
			wantParams:   []any{10, 20},
		},
		{
			name: "multi-field ASC with values",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "col1", Order: postgres.ASC},
					postgres.OrderField{Field: "col2", Order: postgres.DESC},
				},
			},
			cursor:       postgres.CursorPosition{Values: []any{10, 20}},
			paramsOffset: 0,
			order:        postgres.ASC,
			wantStr:      " AND ( col1, col2) > ($1, $2)",
			wantParams:   []any{10, 20},
		},
		{
			name: "empty values returns empty",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			cursor:       postgres.CursorPosition{},
			paramsOffset: 0,
			order:        postgres.DESC,
			wantStr:      "",
			wantParams:   []any{},
		},
		{
			name: "single field with paramsOffset",
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			cursor:       postgres.CursorPosition{Values: []any{99}},
			paramsOffset: 2,
			order:        postgres.ASC,
			wantStr:      " AND id > $3",
			wantParams:   []any{99},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotParams := tt.index.CursorConditionsWithOrder(tt.cursor, tt.paramsOffset, tt.order)
			assert.Equal(t, tt.wantStr, gotStr)
			assert.Equal(t, tt.wantParams, gotParams)
		})
	}
}

func TestGenerateSelectAllOrdered(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		fields  []string
		index   postgres.Index
		limit   uint32
		cursor  postgres.CursorPosition
		order   postgres.Order
		want    *postgres.Query
		wantErr bool
	}{
		{
			name:   "ASC order without cursor",
			table:  "items",
			fields: []string{"id", "name"},
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			limit:  100,
			cursor: postgres.CursorPosition{},
			order:  postgres.ASC,
			want: &postgres.Query{
				QueryString:     `SELECT id, name FROM "items" WHERE true ORDER BY id ASC LIMIT 100`,
				ConditionExists: false,
				Params:          []any{},
			},
		},
		{
			name:   "DESC order without cursor",
			table:  "items",
			fields: []string{"id", "name"},
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			limit:  100,
			cursor: postgres.CursorPosition{},
			order:  postgres.DESC,
			want: &postgres.Query{
				QueryString:     `SELECT id, name FROM "items" WHERE true ORDER BY id DESC LIMIT 100`,
				ConditionExists: false,
				Params:          []any{},
			},
		},
		{
			name:   "DESC order with cursor",
			table:  "items",
			fields: []string{"id", "name"},
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			limit:  50,
			cursor: postgres.CursorPosition{Values: []any{100}},
			order:  postgres.DESC,
			want: &postgres.Query{
				QueryString:     `SELECT id, name FROM "items" WHERE true AND id < $1 ORDER BY id DESC LIMIT 50`,
				ConditionExists: true,
				Params:          []any{100},
			},
		},
		{
			name:   "ASC order with cursor",
			table:  "items",
			fields: []string{"id", "name"},
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			limit:  50,
			cursor: postgres.CursorPosition{Values: []any{100}},
			order:  postgres.ASC,
			want: &postgres.Query{
				QueryString:     `SELECT id, name FROM "items" WHERE true AND id > $1 ORDER BY id ASC LIMIT 50`,
				ConditionExists: true,
				Params:          []any{100},
			},
		},
		{
			name:   "multi-field index DESC with cursor",
			table:  "events",
			fields: []string{"id", "ts", "data"},
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
					postgres.OrderField{Field: "ts", Order: postgres.ASC},
				},
			},
			limit:  20,
			cursor: postgres.CursorPosition{Values: []any{5, 1000}},
			order:  postgres.DESC,
			want: &postgres.Query{
				QueryString:     `SELECT id, ts, data FROM "events" WHERE true AND ( id, ts) < ($1, $2) ORDER BY id DESC, ts DESC LIMIT 20`,
				ConditionExists: true,
				Params:          []any{5, 1000},
			},
		},
		{
			name:   "limit exceeds max",
			table:  "items",
			fields: []string{"id"},
			index: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
				},
			},
			limit:   20000,
			cursor:  postgres.CursorPosition{},
			order:   postgres.ASC,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := postgres.GenerateSelectAllOrdered(tt.table, tt.fields, tt.index, tt.limit, tt.cursor, tt.order)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSelectAllOrdered() =\n  %+v\nwant\n  %+v", got, tt.want)
			}
		})
	}
}
