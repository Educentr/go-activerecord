package postgres_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mailru/activerecord/pkg/activerecord"
	"github.com/mailru/activerecord/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSelect(t *testing.T) {
	type args struct {
		tableName  string
		fieldNames []string
		index      postgres.Index
		keys       [][]any
		offset     uint32
		limit      uint32
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

func TestGenerateUpdate(t *testing.T) {
	tests := []struct {
		name          string
		tableName     string
		primaryIndex  postgres.Index
		updates       []postgres.UpdateParams
		expectedQuery string
		expectedError error
	}{
		{
			name:      "Single update with OpSet",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Ops{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John Doe",
						},
					},
				},
			},
			expectedQuery: `UPDATE "users" SET name = $1  WHERE id  = $2`,
			expectedError: nil,
		},
		{
			name:      "Single update with OpAdd",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Ops{
						{
							Field: "age",
							Op:    activerecord.OpAdd,
							Value: 1,
						},
					},
				},
			},
			expectedQuery: `UPDATE "users" SET age =age + $1  WHERE id  = $2`,
			expectedError: nil,
		},
		{
			name:      "Single update with OpAnd",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Ops{
						{
							Field: "flags",
							Op:    activerecord.OpAnd,
							Value: 1,
						},
					},
				},
			},
			expectedQuery: `UPDATE "users" SET flags =flags & $1  WHERE id  = $2`,
			expectedError: nil,
		},
		{
			name:      "Bulk update not implemented",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Ops{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John Doe",
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Ops{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Jane Doe",
						},
					},
				},
			},
			expectedQuery: "",
			expectedError: fmt.Errorf("bulk update not implemented"),
		},
		{
			name:      "Primary key length mismatch",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					{
						Field: "id",
						Order: postgres.ASC,
					},
					{
						Field: "email",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Ops{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John Doe",
						},
					},
				},
			},
			expectedQuery: "",
			expectedError: fmt.Errorf("primary key length ([1]) not equal to index fields in update 0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := postgres.GenerateUpdate(tt.tableName, tt.primaryIndex, tt.updates)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQuery, q.QueryString)
			}
		})
	}
}

func TestGenerateInsert(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		pk             postgres.Index
		fieldNames     []string
		values         [][]any
		returning      []string
		conflictAction postgres.OnConflictAction
		expectedQuery  string
		expectedParams []any
		expectedErr    error
	}{
		{
			name:      "single insert without conflict",
			tableName: "users",
			pk: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			fieldNames:     []string{"id", "name"},
			values:         [][]any{{1, "John"}},
			returning:      []string{"id"},
			conflictAction: postgres.IgnoreDuplicate,
			expectedQuery:  `INSERT INTO "users" (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id`,
			expectedParams: []any{1, "John"},
			expectedErr:    nil,
		},
		{
			name:      "bulk insert with conflict update",
			tableName: "users",
			pk: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			fieldNames:     []string{"id", "name"},
			values:         [][]any{{1, "John"}, {2, "Doe"}},
			returning:      []string{"id"},
			conflictAction: postgres.UpdateDuplicate,
			expectedQuery:  `INSERT INTO "users" (id, name) VALUES ($1, $2), ($3, $4) ON CONFLICT (id) DO UPDATE SET id=users.id, name=EXCLUDED.name RETURNING id`,
			expectedParams: []any{1, "John", 2, "Doe"},
			expectedErr:    nil,
		},
		{
			name:      "bulk insert with conflict do nothing",
			tableName: "users",
			pk: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			fieldNames:     []string{"id", "name"},
			values:         [][]any{{1, "John"}, {2, "Doe"}},
			returning:      []string{"id"},
			conflictAction: postgres.IgnoreDuplicate,
			expectedQuery:  "",
			expectedParams: nil,
			expectedErr:    fmt.Errorf("can't do bulk insert with 'on_conflict_do_nothing' option"),
		},
		{
			name:      "unknown conflict action",
			tableName: "users",
			pk: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			fieldNames:     []string{"id", "name"},
			values:         [][]any{{1, "John"}},
			returning:      []string{"id"},
			conflictAction: 255,
			expectedQuery:  "",
			expectedParams: nil,
			expectedErr:    fmt.Errorf("unknown conflict action"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := postgres.GenerateInsert(tt.tableName, tt.pk, tt.fieldNames, tt.values, tt.returning, tt.conflictAction)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQuery, q.QueryString)
				assert.Equal(t, tt.expectedParams, q.Params)
			}
		})
	}
}

func TestGenerateDeleteWithRealIndex(t *testing.T) {
	tests := []struct {
		name        string
		tableName   string
		primaryKey  postgres.Index
		keys        [][]any
		expectedSQL string
		expectError bool
	}{
		{
			name:      "single key",
			tableName: "users",
			primaryKey: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			keys:        [][]any{{1}},
			expectedSQL: `DELETE FROM "users" WHERE id  = $1`,
			expectError: false,
		},
		{
			name:      "single key with multiple fields",
			tableName: "users",
			primaryKey: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
					postgres.OrderField{Field: "bla", Order: postgres.ASC},
				},
			},
			keys:        [][]any{{1, 2}},
			expectedSQL: `DELETE FROM "users" WHERE (id, bla)  = ($1, $2)`,
			expectError: false,
		},
		{
			name:      "multiple keys",
			tableName: "users",
			primaryKey: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			keys:        [][]any{{1}, {2}},
			expectedSQL: `DELETE FROM "users" WHERE id  IN ($1, $2)`,
			expectError: false,
		},
		{
			name:      "multiple keys with multiple fields",
			tableName: "users",
			primaryKey: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{
					postgres.OrderField{Field: "id", Order: postgres.ASC},
					postgres.OrderField{Field: "bla", Order: postgres.ASC},
				},
			},
			keys:        [][]any{{1, 2}, {3, 4}},
			expectedSQL: `DELETE FROM "users" WHERE (id, bla)  IN (($1, $2), ($3, $4))`,
			expectError: false,
		},
		{
			name:      "validation error",
			tableName: "users",
			primaryKey: postgres.Index{
				Unique: true,
				Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
			},
			keys:        [][]any{},
			expectedSQL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := postgres.GenerateDelete(tt.tableName, tt.primaryKey, tt.keys)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSQL, q.QueryString)
			}
		})
	}
}
