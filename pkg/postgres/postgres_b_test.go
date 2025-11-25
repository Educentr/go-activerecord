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
					Condition: []postgres.Condition{{Field: "name", Operator: "=", Values: []any{"John"}}},
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
					Condition: []postgres.Condition{{Field: "name", Operator: "=", Values: []any{"John"}}, {Field: "status", Operator: "=", Values: []any{"active"}}},
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
					Condition: []postgres.Condition{{Field: "name", Operator: "=", Values: []any{"John"}}, {Field: "status", Operator: "=", Values: []any{"active", "ready"}}},
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
					Condition: []postgres.Condition{{Field: "name", Operator: "=", Values: []any{"John"}}},
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
		{
			name: "condition_with_greater_than_operator",
			args: args{
				tableName:  "products",
				fieldNames: []string{"id", "name", "price"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "price", Operator: ">", Values: []any{100}}},
				},
				keys:   [][]any{{1}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, price FROM "products" WHERE price > $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{100, 1},
			},
			wantErr: false,
		},
		{
			name: "condition_with_less_than_operator",
			args: args{
				tableName:  "products",
				fieldNames: []string{"id", "name", "price"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "price", Operator: "<", Values: []any{1000}}},
				},
				keys:   [][]any{{5}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, price FROM "products" WHERE price < $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{1000, 5},
			},
			wantErr: false,
		},
		{
			name: "condition_with_greater_equal_operator",
			args: args{
				tableName:  "products",
				fieldNames: []string{"id", "name", "stock"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "stock", Operator: ">=", Values: []any{10}}},
				},
				keys:   [][]any{{3}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, stock FROM "products" WHERE stock >= $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{10, 3},
			},
			wantErr: false,
		},
		{
			name: "condition_with_less_equal_operator",
			args: args{
				tableName:  "products",
				fieldNames: []string{"id", "name", "discount"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "discount", Operator: "<=", Values: []any{50}}},
				},
				keys:   [][]any{{7}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, discount FROM "products" WHERE discount <= $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{50, 7},
			},
			wantErr: false,
		},
		{
			name: "condition_with_not_equal_operator",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "status"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "status", Operator: "!=", Values: []any{"deleted"}}},
				},
				keys:   [][]any{{2}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, status FROM "users" WHERE status != $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{"deleted", 2},
			},
			wantErr: false,
		},
		{
			name: "condition_with_is_null",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "deleted_at"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "deleted_at", Operator: "IS NULL", Values: []any{}}},
				},
				keys:   [][]any{{4}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, deleted_at FROM "users" WHERE deleted_at IS NULL AND id = $1 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{4},
			},
			wantErr: false,
		},
		{
			name: "condition_with_is_not_null",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "verified_at"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "verified_at", Operator: "IS NOT NULL", Values: []any{}}},
				},
				keys:   [][]any{{6}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, verified_at FROM "users" WHERE verified_at IS NOT NULL AND id = $1 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{6},
			},
			wantErr: false,
		},
		{
			name: "condition_with_field_expression_bitwise_and",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "flags"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{FieldExpression: "flags & 1", Operator: "=", Values: []any{1}}},
				},
				keys:   [][]any{{8}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, flags FROM "users" WHERE flags & 1 = $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{1, 8},
			},
			wantErr: false,
		},
		{
			name: "condition_with_field_expression_bitwise_or",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "permissions"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{FieldExpression: "permissions | 4", Operator: "!=", Values: []any{0}}},
				},
				keys:   [][]any{{9}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, permissions FROM "users" WHERE permissions | 4 != $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{0, 9},
			},
			wantErr: false,
		},
		{
			name: "multiple_conditions_with_different_operators",
			args: args{
				tableName:  "products",
				fieldNames: []string{"id", "name", "price", "stock", "deleted_at"},
				index: postgres.Index{
					Unique: false,
					Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{
						{Field: "price", Operator: ">", Values: []any{100}},
						{Field: "stock", Operator: ">=", Values: []any{1}},
						{Field: "deleted_at", Operator: "IS NULL", Values: []any{}},
					},
				},
				keys:   [][]any{{10}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, price, stock, deleted_at FROM "products" WHERE price > $1 AND stock >= $2 AND deleted_at IS NULL AND id = $3 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{100, 1, 10},
			},
			wantErr: false,
		},
		{
			name: "bulk_query_with_comparison_condition",
			args: args{
				tableName:  "products",
				fieldNames: []string{"id", "name", "price"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "price", Operator: ">", Values: []any{50}}},
				},
				keys:   [][]any{{1}, {2}, {3}},
				offset: 0,
				limit:  10,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, price FROM "products" WHERE price > $1 AND id IN ($2, $3, $4) ORDER BY id ASC LIMIT 10`,
				ConditionExists: true,
				Params:          []any{50, 1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "bulk_query_with_null_check_and_comparison",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "age", "deleted_at"},
				index: postgres.Index{
					Unique: false,
					Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{
						{Field: "age", Operator: ">=", Values: []any{18}},
						{Field: "deleted_at", Operator: "IS NULL", Values: []any{}},
					},
				},
				keys:   [][]any{{1}, {2}},
				offset: 0,
				limit:  5,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, age, deleted_at FROM "users" WHERE age >= $1 AND deleted_at IS NULL AND id IN ($2, $3) ORDER BY id ASC LIMIT 5`,
				ConditionExists: true,
				Params:          []any{18, 1, 2},
			},
			wantErr: false,
		},
		{
			name: "complex_condition_with_bitwise_and_null_check",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "flags", "deleted_at"},
				index: postgres.Index{
					Unique: false,
					Fields: postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{
						{FieldExpression: "flags & 1", Operator: "=", Values: []any{1}},
						{Field: "deleted_at", Operator: "IS NULL", Values: []any{}},
					},
				},
				keys:   [][]any{{11}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, flags, deleted_at FROM "users" WHERE flags & 1 = $1 AND deleted_at IS NULL AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{1, 11},
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
		{
			name: "condition_with_boolean_true",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "is_active"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "is_active", Operator: "=", Values: []any{true}}},
				},
				keys:   [][]any{{1}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, is_active FROM "users" WHERE is_active = $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{true, 1},
			},
			wantErr: false,
		},
		{
			name: "condition_with_boolean_false",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "is_deleted"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "is_deleted", Operator: "=", Values: []any{false}}},
				},
				keys:   [][]any{{2}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, is_deleted FROM "users" WHERE is_deleted = $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{false, 2},
			},
			wantErr: false,
		},
		{
			name: "condition_with_boolean_not_equal",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "is_verified"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "is_verified", Operator: "!=", Values: []any{false}}},
				},
				keys:   [][]any{{3}},
				offset: 0,
				limit:  0,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, is_verified FROM "users" WHERE is_verified != $1 AND id = $2 ORDER BY id ASC`,
				ConditionExists: true,
				Params:          []any{false, 3},
			},
			wantErr: false,
		},
		{
			name: "bulk_query_with_boolean_condition",
			args: args{
				tableName:  "users",
				fieldNames: []string{"id", "name", "is_active"},
				index: postgres.Index{
					Unique:    false,
					Fields:    postgres.OrderedFields{postgres.OrderField{Field: "id", Order: postgres.ASC}},
					Condition: []postgres.Condition{{Field: "is_active", Operator: "=", Values: []any{true}}},
				},
				keys:   [][]any{{1}, {2}, {3}},
				offset: 0,
				limit:  10,
				cursor: postgres.CursorPosition{},
			},
			want: &postgres.Query{
				QueryString:     `SELECT id, name, is_active FROM "users" WHERE is_active = $1 AND id IN ($2, $3, $4) ORDER BY id ASC LIMIT 10`,
				ConditionExists: true,
				Params:          []any{true, 1, 2, 3},
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
		expectedParams []any
		expectedError  error
	}{
		{
			name:      "Single update with OpSet",
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
							Value: "John Doe",
						},
					},
				},
			},
			idempotencyKey: []activerecord.FieldValue{},
			expectedQuery:  `UPDATE "users" SET name = $1 WHERE id = $2`,
			expectedParams: []any{"John Doe", 1},
			expectedError:  nil,
		},
		{
			name:      "Single update with OpAdd",
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
			expectedParams: []any{1, "John Doe", 2, "Jane Doe"},
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
	}


	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := postgres.GenerateUpdate(tt.tableName, tt.primaryIndex, tt.updates, tt.idempotencyKey)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQuery, q.QueryString)
				if tt.expectedParams != nil {
					assert.Equal(t, tt.expectedParams, q.Params, "Query params")
				}
			}
		})
	}
}

func TestGenerateBulkUpdateClustered(t *testing.T) {
	tests := []struct {
		name             string
		tableName        string
		primaryIndex     postgres.Index
		updates          []postgres.UpdateParams
		idempotencyKey   []activerecord.FieldValue
		expectedClusters int
		expectedSizes    []int
		expectedFields   [][]string
		expectedParams   [][]any // Ожидаемые params для каждого кластера
		expectedError    error
	}{
		{
			name:      "Single cluster - all objects update same fields",
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
			},
			idempotencyKey:   []activerecord.FieldValue{},
			expectedClusters: 1,
			expectedSizes:    []int{2},
			expectedFields:   [][]string{{"name"}},
			expectedParams:   [][]any{{1, "John", 2, "Jane"}},
			expectedError:    nil,
		},
		{
			name:      "Multiple clusters - objects update different fields",
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
					},
				},
			},
			idempotencyKey:   []activerecord.FieldValue{},
			expectedClusters: 2,
			expectedSizes:    []int{2, 1},
			expectedFields:   [][]string{{"counter"}, {"name"}},
			expectedParams:   [][]any{{1, 5, 3, 10}, {2, "Jane"}},
			expectedError:    nil,
		},
		{
			name:      "Multiple clusters - different field sets",
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
							Value: "Alice",
						},
						{
							Field: "email",
							Op:    activerecord.OpSet,
							Value: "alice@example.com",
						},
					},
				},
				{
					PK: []any{2},
					Ops: []postgres.Operation{
						{
							Field: "counter",
							Op:    activerecord.OpAdd,
							Value: 1,
						},
					},
				},
				{
					PK: []any{3},
					Ops: []postgres.Operation{
						{
							Field: "name",
							Op:    activerecord.OpSet,
							Value: "Bob",
						},
						{
							Field: "email",
							Op:    activerecord.OpSet,
							Value: "bob@example.com",
						},
					},
				},
			},
			idempotencyKey:   []activerecord.FieldValue{},
			expectedClusters: 2,
			expectedSizes:    []int{1, 2},
			expectedFields:   [][]string{{"counter"}, {"email", "name"}},
			expectedParams:   [][]any{{2, 1}, {1, "alice@example.com", "Alice", 3, "bob@example.com", "Bob"}},
			expectedError:    nil,
		},
		{
			name:      "Type casting in VALUES - verify SQL structure",
			tableName: "payments",
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
					PK: []any{int64(100)},
					Ops: []postgres.Operation{
						{
							Field: "status",
							Op:    activerecord.OpSet,
							Value: "completed",
						},
					},
				},
				{
					PK: []any{int64(200)},
					Ops: []postgres.Operation{
						{
							Field: "status",
							Op:    activerecord.OpSet,
							Value: "pending",
						},
					},
				},
			},
			idempotencyKey:   []activerecord.FieldValue{},
			expectedClusters: 1,
			expectedSizes:    []int{2},
			expectedFields:   [][]string{{"status"}},
			expectedParams:   [][]any{{int64(100), "completed", int64(200), "pending"}},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём мапу типов для тестов
			fieldTypes := postgres.FieldTypeMap{
				"id":      "BIGINT",
				"name":    "VARCHAR(n)",
				"email":   "VARCHAR(n)",
				"counter": "INTEGER",
				"status":  "VARCHAR(50)",
			}
			result, err := postgres.GenerateBulkUpdateClustered(tt.tableName, tt.primaryIndex, tt.updates, tt.idempotencyKey, fieldTypes)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedClusters, len(result.Queries), "Number of clusters")
				assert.Equal(t, tt.expectedSizes, result.ClusterSizes, "Cluster sizes")
				assert.Equal(t, tt.expectedFields, result.ClusterFields, "Cluster fields")

				// Проверяем params для каждого кластера
				for i, expectedParams := range tt.expectedParams {
					assert.Equal(t, expectedParams, result.Queries[i].Params, "Params for cluster %d", i)
				}

				// Для теста с type casting проверяем что в первой строке VALUES есть ::type
				if tt.name == "Type casting in VALUES - verify SQL structure" {
					query := result.Queries[0].QueryString
					assert.Contains(t, query, "$1::BIGINT", "First row should have type cast for id")
					assert.Contains(t, query, "$2::VARCHAR(50)", "First row should have type cast for status")
					// Вторая строка не должна иметь type casts
					assert.Contains(t, query, "($3, $4)", "Second row should not have type casts")
				}
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
