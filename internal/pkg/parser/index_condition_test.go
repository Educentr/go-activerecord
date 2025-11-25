package parser

import (
	"testing"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestParseConditionValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string with single quotes",
			input:    "''",
			expected: []string{""},
		},
		{
			name:     "empty string with double quotes",
			input:    `""`,
			expected: []string{""},
		},
		{
			name:     "single value without quotes",
			input:    "1",
			expected: []string{"1"},
		},
		{
			name:     "single value with single quotes",
			input:    "'active'",
			expected: []string{"active"},
		},
		{
			name:     "single value with double quotes",
			input:    `"active"`,
			expected: []string{"active"},
		},
		{
			name:     "multiple values without quotes",
			input:    "1,2,3",
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "multiple values with quotes",
			input:    "'active','pending','completed'",
			expected: []string{"active", "pending", "completed"},
		},
		{
			name:     "mixed empty and non-empty strings",
			input:    "'','active',''",
			expected: []string{"", "active", ""},
		},
		{
			name:     "values with spaces",
			input:    " '' , 'active' , 'pending' ",
			expected: []string{"", "active", "pending"},
		},
		{
			name:     "empty string followed by value",
			input:    "'',1",
			expected: []string{"", "1"},
		},
		{
			name:     "value followed by empty string",
			input:    "1,''",
			expected: []string{"1", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConditionValues(tt.input)
			assert.Check(t, cmp.DeepEqual(tt.expected, result), "Invalid result for test '%s'", tt.name)
		})
	}
}

func TestParseIndexConditionTag(t *testing.T) {
	tests := []struct {
		name      string
		condTag   string
		fieldsMap map[string]int
		want      map[int]ds.IndexCondition
		wantErr   bool
	}{
		{
			name:      "empty string condition with single quotes",
			condTag:   "Status[=]''",
			fieldsMap: map[string]int{"Status": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "=",
					Value:         []string{""},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "empty string condition with double quotes",
			condTag:   `Status[=]""`,
			fieldsMap: map[string]int{"Status": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "=",
					Value:         []string{""},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "not equal empty string",
			condTag:   "Status[!=]''",
			fieldsMap: map[string]int{"Status": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "!=",
					Value:         []string{""},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "multiple conditions including empty string",
			condTag:   "Status[=]'';Type[!=]''",
			fieldsMap: map[string]int{"Status": 0, "Type": 1},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "=",
					Value:         []string{""},
					IsNullCheck:   false,
				},
				1: {
					ConditionType: "!=",
					Value:         []string{""},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "regular value with quotes",
			condTag:   "Status[=]'active'",
			fieldsMap: map[string]int{"Status": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "=",
					Value:         []string{"active"},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "is null condition",
			condTag:   "Status[is null]",
			fieldsMap: map[string]int{"Status": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "is null",
					Value:         []string{},
					IsNullCheck:   true,
				},
			},
			wantErr: false,
		},
		{
			name:      "boolean true condition",
			condTag:   "IsActive[=]true",
			fieldsMap: map[string]int{"IsActive": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "=",
					Value:         []string{"true"},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "boolean false condition",
			condTag:   "IsActive[=]false",
			fieldsMap: map[string]int{"IsActive": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "=",
					Value:         []string{"false"},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
		{
			name:      "boolean not equal true",
			condTag:   "IsActive[!=]true",
			fieldsMap: map[string]int{"IsActive": 0},
			want: map[int]ds.IndexCondition{
				0: {
					ConditionType: "!=",
					Value:         []string{"true"},
					IsNullCheck:   false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIndexConditionTag(tt.condTag, tt.fieldsMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIndexConditionTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Check(t, cmp.DeepEqual(tt.want, got), "Invalid result for test '%s'", tt.name)
			}
		})
	}
}
