package postgres

import (
	"fmt"
	"strings"
)

type Order uint8

const (
	ASC Order = iota
	DESC
)

type CursorPosition struct {
	Values []any
	Order  Order
}

type OrderField struct {
	Field string
	Order Order
}

func (of OrderField) String() string {
	str := of.Field

	switch of.Order {
	case ASC:
		str += " ASC"
	case DESC:
		str += " DESC"
	}

	return str
}

type OrderedFields []OrderField

func (ofs OrderedFields) GetFieldNames() []string {
	retFields := make([]string, 0, len(ofs))
	for _, f := range ofs {
		retFields = append(retFields, f.Field)
	}

	return retFields
}

type Index struct {
	Fields       OrderedFields
	Unique       bool
	Condition    []string
	DefaultLimit uint16
}

func (i Index) MultiField() bool {
	return len(i.Fields) > 1
}

func (i Index) OrderConditions() string {
	orderFields := []string{}

	for _, f := range i.Fields {
		orderFields = append(orderFields, f.String())
	}

	return " ORDER BY " + strings.Join(orderFields, ", ")
}

func (i Index) CursorConditions(c CursorPosition, paramsOffset int) (string, []any) {
	str := ""
	params := []any{}

	if len(c.Values) == 0 {
		return str, params
	}

	if i.MultiField() {
		str += " AND ( " + strings.Join(i.Fields.GetFieldNames(), ", ") + ") > ("

		placeholder := make([]string, 0, len(c.Values))
		for _, b := range c.Values {
			params = append(params, b)
			placeholder = append(placeholder, fmt.Sprintf("$%d", len(params)+paramsOffset))
		}

		str += strings.Join(placeholder, ", ") + ")"
	} else {
		params = append(params, c.Values[0])
		str += fmt.Sprintf(" AND %s > $%d", i.Fields[0].Field, len(params)+paramsOffset)
	}

	return str, params
}

func (i Index) ValidateCursor(c CursorPosition) error {
	if len(c.Values) == 0 {
		return nil
	}

	if len(c.Values) != len(i.Fields) {
		return fmt.Errorf("cursor length not equal to index fields")
	}

	return nil
}

func (i Index) validateKeys(keys [][]any) error {
	if len(keys) == 0 {
		return fmt.Errorf("empty keys")
	}

	keyCount := len(keys[0])

	for _, k := range keys {
		if len(k) > len(i.Fields) {
			return fmt.Errorf("not many field keys")
		}

		if keyCount != len(k) {
			return fmt.Errorf("different key count not allowed")
		}
	}

	return nil
}

func (i Index) Conditions() string {
	if i.Condition != nil && len(i.Condition) > 0 {
		return strings.Join(i.Condition, " AND ") + " AND "
	}

	return ""
}

func (i Index) ConditionFields() string {
	if i.MultiField() {
		return "(" + strings.Join(i.Fields.GetFieldNames(), ", ") + ")"
	}

	return i.Fields[0].Field
}

func (i Index) GenerateWhereKeys(q *Query, keys [][]any) {
	if len(keys) > 1 {
		placeholders := make([]string, 0, len(keys))
		if i.MultiField() {
			for _, key := range keys {
				innerPlaceholder := make([]string, 0, len(key))

				for _, kField := range key {
					innerPlaceholder = append(innerPlaceholder, fmt.Sprintf("$%d", q.AddParams(kField)))
				}

				placeholders = append(placeholders, "("+strings.Join(innerPlaceholder, ", ")+")")
			}
		} else {
			for _, key := range keys {
				placeholders = append(placeholders, fmt.Sprintf("$%d", q.AddParams(key[0])))
			}
		}

		q.QueryString += " IN (" + strings.Join(placeholders, ", ") + ")"
	} else {
		if i.MultiField() {
			innerPlaceholder := make([]string, 0, len(keys[0]))

			for _, kField := range keys[0] {
				innerPlaceholder = append(innerPlaceholder, fmt.Sprintf("$%d", q.AddParams(kField)))
			}

			q.QueryString += " = (" + strings.Join(innerPlaceholder, ", ") + ")"
		} else {
			q.QueryString += fmt.Sprintf(" = $%d", q.AddParams(keys[0][0]))
		}
	}

}
