package postgres_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Educentr/go-activerecord/v3/pkg/activerecord"
	"github.com/Educentr/go-activerecord/v3/pkg/postgres"
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
				QueryString:     `SELECT id, name, email FROM "users" WHERE id = $1`,
				ConditionExists: true,
				Params:          []any{1},
			},
			wantErr: false,
		},
		{
			name: "simple with index condition",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "email"},
				index: postgres.Index{
					Unique:    true,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "name", Values: []any{"John"}}},
				},
				keys:   [][]any{{1}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, email FROM "users" WHERE name = $1 AND id = $2`,
				ConditionExists: true,
				Params:          []any{"John", 1},
			},
			wantErr: false,
		},
		{
			name: "simple with multi field index condition",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "email", "status"},
				index: postgres.Index{
					Unique:    true,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "name", Values: []any{"John"}}, {Field: "status", Values: []any{"active"}}},
				},
				keys:   [][]any{{1}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, email, status FROM "users" WHERE name = $1 AND status = $2 AND id = $3`,
				ConditionExists: true,
				Params:          []any{"John", "active", 1},
			},
			wantErr: false,
		},
		{
			name: "simple with multi field bulk index condition",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "email", "status"},
				index: postgres.Index{
					Unique:    true,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "name", Values: []any{"John"}}, {Field: "status", Values: []any{"active", "ready"}}},
				},
				keys:   [][]any{{1}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, email, status FROM "users" WHERE name = $1 AND status IN ($2, $3) AND id = $4`,
				ConditionExists: true,
				Params:          []any{"John", "active", "ready", 1},
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
				QueryString:     `SELECT id, name, email FROM "users" WHERE id IN ($1, $2, $3) ORDER BY id ASC LIMIT 10`,
				ConditionExists: true,
				Params:          []any{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "bulk with index condition",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "email"},
				index: postgres.Index{
					Unique:    true,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "name", Values: []any{"John"}}},
				},
				keys:   [][]any{{1}, {2}, {3}},
				offset: 0,
				limit:  10,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, email FROM "users" WHERE name = $1 AND id IN ($2, $3, $4) ORDER BY id ASC LIMIT 10`,
				ConditionExists: true,
				Params:          []any{"John", 1, 2, 3},
			},
			wantErr: false,
		},
		// ToDo test for multi field index and bulk query
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
				QueryString:     `SELECT id, parent, name, email FROM "users" WHERE (id, parent) IN (($1, $2), ($3, $4), ($5, $6)) ORDER BY id ASC, parent DESC LIMIT 5 OFFSET 2`,
				ConditionExists: true,
				Params:          []any{1, 1, 2, 1, 3, 4},
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
		name           string
		tableName      string
		primaryIndex   postgres.Index
		updates        []postgres.UpdateParams
		idempotencyKey []activerecord.FieldValue
		expectedQuery  string
		expectedError  error
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
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John Doe",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  `UPDATE "users" SET name = $1 WHERE id = $2`,
			expectedError:  nil,
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
					Ops: []postgres.Operation{
						{
							Field: "age",
							Op:    activerecord.OpAdd,
							Value: 1,
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  `UPDATE "users" SET age = age + $1 WHERE id = $2 RETURNING age`,
			expectedError:  nil,
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
					Ops: []postgres.Operation{
						{
							Field: "flags",
							Op:    activerecord.OpAnd,
							Value: 1,
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  `UPDATE "users" SET flags = flags & $1 WHERE id = $2 RETURNING flags`,
			expectedError:  nil,
		},
		{
			name:      "Update with DBSerializer",
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
					Ops: []postgres.Operation{
						{
							Field: "Date",
							Op:    activerecord.OpSet,
							Value: 1,
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  `UPDATE "users" SET Date = $1 WHERE id = $2`,
			expectedError:  nil,
		},
		{
			name:      "Bulk update with single field",
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
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John Doe",
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Jane Doe",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery: `UPDATE users AS t
SET name = v.name
FROM (VALUES
    ($1, $2),
    ($3, $4)
) AS v(id, name)
WHERE t.id = v.id`,
			expectedError: nil,
		},
		{
			name:      "Primary key length mismatch",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
					postgres.OrderField{
						Field: "email",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John Doe",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  "",
			expectedError:  fmt.Errorf("primary key length ([1]) not equal to index fields in update 0"),
		},
		{
			name:      "Bulk update with OpAdd - mutating operation",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 5,
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 10,
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery: `UPDATE users AS t
SET counter = t.counter + v.counter
FROM (VALUES
    ($1, $2),
    ($3, $4)
) AS v(id, counter)
WHERE t.id = v.id
RETURNING t.id, t.counter`,
			expectedError: nil,
		},
		{
			name:      "Bulk update with OpAnd - mutating operation",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "flags",
							Op:    activerecord.OpAnd,
							Value: 0xFF,
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "flags",
							Op:    activerecord.OpAnd,
							Value: 0x0F,
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery: `UPDATE users AS t
SET flags = t.flags & v.flags
FROM (VALUES
    ($1, $2),
    ($3, $4)
) AS v(id, flags)
WHERE t.id = v.id
RETURNING t.id, t.flags`,
			expectedError: nil,
		},
		{
			name:      "Bulk update with mixed operations should fail",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 5,
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpSet,
							Value: 10,
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  "",
			expectedError:  fmt.Errorf("field counter uses multiple operation types in bulk update, which is not supported"),
		},
		{
			name:      "Bulk update with multiple fields including mutating op",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 5,
						},
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John",
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 10,
						},
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Jane",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery: `UPDATE users AS t
SET counter = t.counter + v.counter, name = v.name
FROM (VALUES
    ($1, $2, $3),
    ($4, $5, $6)
) AS v(id, counter, name)
WHERE t.id = v.id
RETURNING t.id, t.counter`,
			expectedError: nil,
		},
		{
			name:      "Bulk update with only OpSet - no RETURNING needed",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "John",
						},
						{
							Field: "email",
							Op:    activerecord.OpSet,
							Value: "john@example.com",
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Jane",
						},
						{
							Field: "email",
							Op:    activerecord.OpSet,
							Value: "jane@example.com",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery: `UPDATE users AS t
SET email = v.email, name = v.name
FROM (VALUES
    ($1, $2, $3),
    ($4, $5, $6)
) AS v(id, email, name)
WHERE t.id = v.id`,
			expectedError: nil,
		},
		{
			name:      "Bulk update with partial mutations - only some objects mutate field",
			tableName: "users",
			primaryIndex: postgres.Index{
				Fields: postgres.OrderedFields{
					postgres.OrderField{
						Field: "id",
						Order: postgres.ASC,
					},
				},
				Unique: true,
			},
			updates: []postgres.UpdateParams{
				{
					PK: []any{1},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 5,
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Jane",
						},
					},
				},
				{
					PK: []any{3},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 10,
						},
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Bob",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery: `UPDATE users AS t
SET counter = t.counter + v.counter, name = v.name
FROM (VALUES
    ($1, $2, NULL),
    ($3, NULL, $4),
    ($5, $6, $7)
) AS v(id, counter, name)
WHERE t.id = v.id
RETURNING t.id, t.counter`,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := postgres.GenerateUpdate(tt.tableName, tt.primaryIndex, tt.updates, tt.idempotencyKey)
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
			conflictAction: postgres.Replace,
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
			q, err := postgres.GenerateInsert(tt.tableName, tt.pk, tt.fieldNames, tt.values, tt.returning, tt.conflictAction, tt.pk)

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
			expectedSQL: `DELETE FROM "users" WHERE id = $1`,
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
			expectedSQL: `DELETE FROM "users" WHERE (id, bla) = ($1, $2)`,
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
			expectedSQL: `DELETE FROM "users" WHERE id IN ($1, $2)`,
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
			expectedSQL: `DELETE FROM "users" WHERE (id, bla) IN (($1, $2), ($3, $4))`,
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
