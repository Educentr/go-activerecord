package postgres_test

import (
	"reflect"
	"testing"

	"github.com/Educentr/go-activerecord/v3/pkg/postgres"
)

// Условие индекса со списком значений должно давать IN/NOT IN: такой же
// предикат argen пишет в DDL частичного индекса, иначе планировщик не применит
// индекс к запросу
func TestQuery_AddCondition(t *testing.T) {
	tests := []struct {
		name       string
		condition  postgres.Condition
		wantQuery  string
		wantParams []any
	}{
		{
			name:       "single value",
			condition:  postgres.Condition{Field: "mid", Operator: "=", Values: []any{int32(1)}},
			wantQuery:  "mid = $1",
			wantParams: []any{int32(1)},
		},
		{
			name:       "multiple values with equality",
			condition:  postgres.Condition{Field: "mid", Operator: "=", Values: []any{int32(1), int32(2), int32(3)}},
			wantQuery:  "mid IN ($1, $2, $3)",
			wantParams: []any{int32(1), int32(2), int32(3)},
		},
		{
			name:       "multiple values with inequality",
			condition:  postgres.Condition{Field: "status", Operator: "!=", Values: []any{"draft", "deleted"}},
			wantQuery:  "status NOT IN ($1, $2)",
			wantParams: []any{"draft", "deleted"},
		},
		{
			name:       "multiple values with field expression",
			condition:  postgres.Condition{FieldExpression: "flags&1", Operator: "=", Values: []any{int64(0), int64(1)}},
			wantQuery:  "flags&1 IN ($1, $2)",
			wantParams: []any{int64(0), int64(1)},
		},
		{
			name:       "null check without values",
			condition:  postgres.Condition{Field: "deleted_at", Operator: "IS NULL"},
			wantQuery:  "deleted_at IS NULL",
			wantParams: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &postgres.Query{Params: []any{}}

			q.AddCondition(tt.condition)

			if q.QueryString != tt.wantQuery {
				t.Errorf("QueryString = %q, want %q", q.QueryString, tt.wantQuery)
			}

			if !reflect.DeepEqual(q.Params, tt.wantParams) {
				t.Errorf("Params = %v, want %v", q.Params, tt.wantParams)
			}
		})
	}
}

// Список значений поддерживают только = и !=; для остальных операторов его
// нечем записать, argen такое условие отвергает на генерации
func TestQuery_AddCondition_MultipleValuesUnsupportedOperator(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for operator '>' with multiple values")
		}
	}()

	q := &postgres.Query{Params: []any{}}

	q.AddCondition(postgres.Condition{Field: "mid", Operator: ">", Values: []any{int32(1), int32(2)}})
}

func TestConditionListOperator(t *testing.T) {
	tests := []struct {
		operator string
		want     string
		wantEx   bool
	}{
		{operator: "=", want: "IN", wantEx: true},
		{operator: "!=", want: "NOT IN", wantEx: true},
		{operator: ">", want: "", wantEx: false},
		{operator: "<=", want: "", wantEx: false},
		{operator: "IS NULL", want: "", wantEx: false},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			got, ex := postgres.ConditionListOperator(tt.operator)
			if got != tt.want || ex != tt.wantEx {
				t.Errorf("ConditionListOperator(%q) = (%q, %v), want (%q, %v)", tt.operator, got, ex, tt.want, tt.wantEx)
			}
		})
	}
}
