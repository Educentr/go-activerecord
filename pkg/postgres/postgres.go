package postgres

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Educentr/go-activerecord/v3/pkg/activerecord"
)

// type WhereCondition

// ToDo merge with octopus InsertModeInserOrReplace, e.t.c.
const (
	Replace OnConflictAction = iota
	IgnoreDuplicate
	NoDuplicateAction
)

type DefaultKeyword bool

const DefaultValueDB DefaultKeyword = true

type QueryBuilderState uint8

const (
	QueryBuilderStateWhere QueryBuilderState = iota
	QueryBuilderStateOrderBy
	QueryBuilderStateLimit
	QueryBuilderStateOffset
)

type Query struct {
	QueryString     string
	ConditionExists bool // Признак того, что уже есть условие в запросе
	Params          []any
	// state       QueryBuilderState
}

func NewSelectQuery(tableName string, fieldNames []string, i Index) *Query {
	// ToDo quote field names
	q := &Query{
		QueryString: fmt.Sprintf(`SELECT %s FROM %s`,
			strings.Join(fieldNames, ", "),
			QuoteIdentifier(tableName),
		),
		Params: []any{},
	}

	q.AddWhereBlock()

	if len(i.Condition) > 0 {
		for _, c := range i.Condition {
			q.AddWhereQuery(c.Field)
			q.GenerateWhereKeys(false, c.GetValues())
		}
	}

	q.ConditionFields(i.Fields.GetFieldNames())

	return q
}

func NewUpdateQuery(tableName string) *Query {
	return &Query{
		QueryString: fmt.Sprintf(`UPDATE %s SET `,
			QuoteIdentifier(tableName),
		),
		Params: []any{},
	}
}

func NewDeleteQuery(tableName string, pk Index) *Query {
	q := &Query{
		QueryString: fmt.Sprintf(`DELETE FROM %s`,
			QuoteIdentifier(tableName),
		),
		Params: []any{},
	}

	q.AddWhereBlock()
	q.ConditionFields(pk.Fields.GetFieldNames())

	return q
}

func NewInsertQuery(tableName string, fieldNames []string) *Query {
	q := &Query{
		QueryString: fmt.Sprintf(`INSERT INTO %s (%s) VALUES `,
			QuoteIdentifier(tableName),
			strings.Join(fieldNames, ", "),
		),
	}

	return q
}

func (q *Query) AddReturning(fieldNames []string) {
	if len(fieldNames) == 0 {
		return
	}

	q.QueryString += fmt.Sprintf(" RETURNING %s",
		strings.Join(fieldNames, ", "),
	)
}

func (q *Query) GenerateWhereKeys(multiField bool, keys [][]any) {
	if len(keys) > 1 {
		placeholders := make([]string, 0, len(keys))
		if multiField {
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
		if multiField {
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

func (q *Query) AddOnConflictDoNothing(fieldNames []string) {
	q.QueryString += " ON CONFLICT DO NOTHING"
}

func (q *Query) AddOnConflictDoUpdate(tableName string, pk Index, fieldNames []string, conflictKey Index) {
	pkfields := make(map[string]struct{}, len(conflictKey.Fields))
	updateFields := []string{}

	for _, pkf := range conflictKey.Fields.GetFieldNames() {
		updateFields = append(updateFields, fmt.Sprintf("%s=%s.%s", pkf, tableName, pkf))
		pkfields[pkf] = struct{}{}
	}

	for _, f := range fieldNames {
		if _, ex := pkfields[f]; ex {
			continue
		}

		updateFields = append(updateFields, fmt.Sprintf("%s=EXCLUDED.%s", f, f))
	}

	q.QueryString += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(conflictKey.Fields.GetFieldNames(), ", "),
		strings.Join(updateFields, ", "),
	)
}

func (q *Query) AddFieldValue(fv ...activerecord.FieldValue) {
	for _, f := range fv {
		q.QueryString += fmt.Sprintf("%s = $%d AND ", f.Field, q.AddParams(f.Value))
	}
}

func (q *Query) AddParams(key ...any) int {
	q.Params = append(q.Params, key...)

	return len(q.Params)
}

func (q *Query) AddWhereQuery(cond string) {
	if q.ConditionExists {
		q.QueryString += " AND "
	}

	q.ConditionExists = true

	q.AddQuery(cond)
}

func (q *Query) AddQuery(cond string) {
	if cond == "" {
		panic("empty condition")
	}

	q.QueryString += cond
}

func (q *Query) AddWhereCondition(cond string, key []any) {
	if (len(key) == 0) != (cond == "") {
		panic("empty condition or key")
	}

	if cond == "" {
		return
	}

	q.AddParams(key...)
	q.AddWhereQuery(cond)
}

func (q *Query) AddSetStatement(field string, key any) {
	q.AddQuery(field + " = $" + fmt.Sprintf("%d", key))
}

func (q *Query) AddWhereBlock(cond ...string) {
	q.QueryString += " WHERE "
	for _, c := range cond {
		q.AddWhereQuery(c)
	}
}

func (q *Query) AddLimitOffset(limit, offset uint32) {
	if limit > 0 {
		q.QueryString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if offset > 0 {
		q.QueryString += fmt.Sprintf(" OFFSET %d", offset)
	}
}

func QuoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func GenerateSelectAll(tableName string, fieldNames []string, index Index, limit uint32, cursor CursorPosition) (*Query, error) {
	// ToDo quote field names
	q := &Query{
		QueryString: fmt.Sprintf("SELECT %s FROM %s WHERE true", strings.Join(fieldNames, ", "), QuoteIdentifier(tableName)),
		Params:      []any{},
	}

	if limit > uint32(MaxLimit) {
		return nil, fmt.Errorf("limit %d is more than max limit %d", limit, MaxLimit)
	}

	q.AddWhereCondition(index.CursorConditions(cursor, len(q.Params)))
	q.AddQuery(index.OrderConditions())

	q.AddLimitOffset(limit, 0)

	return q, nil
}

func GenerateSelect(tableName string, fieldNames []string, index Index, keys [][]any, offset, limit uint32, cursor CursorPosition) (*Query, error) {
	if err := index.validateKeys(keys); err != nil {
		return nil, err
	}

	if err := index.ValidateCursor(cursor); err != nil {
		return nil, err
	}

	bulkSelect := len(keys) > 1
	oneRowResult := !bulkSelect && index.Unique

	q := NewSelectQuery(tableName, fieldNames, index)

	// ToDo work with array fields
	// $pg_request->{query} .= ' && $' . $i++;
	// push @{$pg_request->{params}}, '{' . MR::Pg->encode_array_field($request->{keys}) . '}';

	q.GenerateWhereKeys(index.MultiField(), keys)

	q.AddWhereCondition(index.CursorConditions(cursor, len(q.Params)-1))

	if !oneRowResult {
		q.AddQuery(index.OrderConditions())
	}

	q.AddLimitOffset(limit, offset)

	// ToDo add additional conditions
	// if($opts{order} || $opts{condition}) {
	// 	$pg_request->{query} = 'SELECT * FROM ('.$pg_request->{query}.') AS J';

	// 	if ($opts{condition} && @{$opts{condition}}) {
	// 		my @condition_where;
	// 		foreach my $item (@{$opts{condition}}) {
	// 			my ($field, $value) = @{$item}{'field', 'value'};
	// 			if($array_fields{$field}) {
	// 				push @condition_where, $field . ' && $' . $i++;
	// 				push @{$pg_request->{params}}, '{' . MR::Pg->encode_array_field(ref $value ? $value : [$value]) . '}';
	// 			} elsif(ref($value) eq 'ARRAY') {
	// 				push @condition_where, $field_sql_deserialized_name{$field} . ' IN (?)';
	// 				push @{$pg_request->{params_in}}, $value;
	// 			} else {
	// 				push @condition_where, $field_sql_deserialized_name{$field} . ' = $' . $i++;
	// 				push @{$pg_request->{params}}, $value;
	// 			}
	// 		}
	// 		$pg_request->{query} .=  ' WHERE '. (join ' AND ', @condition_where);
	// 	}

	// }
	// confess 'Wrong order for response count' if $opts{order} && !$opts{condition} && $request->{limit} && $request->{limit} == $response_count;

	// $pg_request->{use_replica} = delete $opts{use_replica} if exists $opts{use_replica}; #TODO up replica flag on object
	// my ($response_count, $response) = $db_class->selectall_arrayref($db, $pg_request);

	// my $result = $class->$select_response($response, %resp_opts);
	// $result = $self->select_postprocess($result, $index, $keys, %opts);
	// return $result;

	return q, nil
}

// ToDo заменить idempotencyKey передаваемые явно в функцию, на With...
// Но возможно это и не понадобиться потому, что надо много переделывать. Нужно отдельно собирать, все что должны сделать
// все With... и потом понимать в какое место запроса надо положить результат.
// type QueryOption interface {
// 	apply(*Query) error
// }

// type optionQueryFunc func(*Query) error

// func (o optionQueryFunc) apply(c *Query) error {
// 	return o(c)
// }

// func WithIdempotencyKey(idempotencyKey []FieldValue) QueryOption {
// 	return optionQueryFunc(func(q *Query) error {
// 		q.AddFieldValue(idempotencyKey...)

// 		return nil
// 	})

// }

func GenerateUpdate(tableName string, primaryIndex Index, updates []UpdateParams, idempotencyKey []activerecord.FieldValue) (*Query, error) {
	isBulk := len(updates) > 1

	if isBulk {
		return generateBulkUpdate(tableName, primaryIndex, updates, idempotencyKey)
	}

	q := NewUpdateQuery(tableName)

	for num, u := range updates {
		if len(u.PK) != len(primaryIndex.Fields) {
			return nil, fmt.Errorf("primary key length (%+v) not equal to index fields in update %d", u.PK, num)
		}

		returning := []string{}

		for n, op := range u.Ops {
			// ToDo sql serializers
			operation := ""
			if n > 0 {
				operation += ", "
			}

			operation += op.Field + " = "

			sql := "$%d"
			ret := op.Field

			switch op.Op {
			case activerecord.OpSet:
				operation += fmt.Sprintf(sql, q.AddParams(op.Value))
			case activerecord.OpAdd:
				operation += op.Field + " + " + fmt.Sprintf(sql, q.AddParams(op.Value))
				returning = append(returning, ret)
			case activerecord.OpAnd:
				operation += op.Field + " & " + fmt.Sprintf(sql, q.AddParams(op.Value))
				returning = append(returning, ret)
			default:
				return nil, fmt.Errorf("unknown operation %d or not implemented", op.Op)
			}

			q.AddQuery(operation)
		}

		q.AddWhereBlock()

		q.AddFieldValue(idempotencyKey...)

		q.ConditionFields(primaryIndex.Fields.GetFieldNames())

		if primaryIndex.MultiField() {
			innerPlaceholder := make([]string, 0, len(u.PK))

			for _, kField := range u.PK {
				innerPlaceholder = append(innerPlaceholder, fmt.Sprintf("$%d", q.AddParams(kField)))
			}

			q.QueryString += " = (" + strings.Join(innerPlaceholder, ", ") + ")"
		} else {
			q.QueryString += fmt.Sprintf(" = $%d", q.AddParams(u.PK[0]))
		}

		if len(returning) > 0 {
			q.AddReturning(returning)
		}
	}

	return q, nil
}

// generateBulkUpdate creates UPDATE ... FROM (VALUES ...) query for bulk updates
func generateBulkUpdate(tableName string, primaryIndex Index, updates []UpdateParams, idempotencyKey []activerecord.FieldValue) (*Query, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("empty updates")
	}

	if len(idempotencyKey) > 0 {
		return nil, fmt.Errorf("idempotency keys not supported in bulk update")
	}

	// Собираем все уникальные поля из всех UpdateOps
	allFields := make(map[string]bool)
	for _, u := range updates {
		for _, op := range u.Ops {
			allFields[op.Field] = true
		}
	}

	if len(allFields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Проверяем что все PK одинаковой длины
	pkLen := len(primaryIndex.Fields)
	for i, u := range updates {
		if len(u.PK) != pkLen {
			return nil, fmt.Errorf("primary key length mismatch in update %d", i)
		}
	}

	// Создаем упорядоченный список полей для обновления
	fieldsList := make([]string, 0, len(allFields))
	for field := range allFields {
		fieldsList = append(fieldsList, field)
	}

	// Сортируем для стабильности
	sort.Strings(fieldsList)

	q := &Query{
		QueryString: "",
		Params:      []any{},
	}

	// UPDATE table_name AS t
	q.QueryString = fmt.Sprintf("UPDATE %s AS t\nSET ", tableName)

	// SET field1 = v.field1, field2 = v.field2, ...
	setFields := make([]string, 0, len(fieldsList))
	for _, field := range fieldsList {
		setFields = append(setFields, fmt.Sprintf("%s = v.%s", field, field))
	}
	q.QueryString += strings.Join(setFields, ", ")

	// FROM (VALUES ...)
	q.QueryString += "\nFROM (VALUES\n"

	// Генерируем строки VALUES
	valueRows := make([]string, 0, len(updates))
	for _, u := range updates {
		// Создаем map для быстрого поиска значений
		opsMap := make(map[string]Operation)
		for _, op := range u.Ops {
			opsMap[op.Field] = op
		}

		// Собираем значения: сначала PK, потом поля в порядке fieldsList
		values := make([]string, 0, pkLen+len(fieldsList))

		// PK values
		for _, pkVal := range u.PK {
			values = append(values, fmt.Sprintf("$%d", q.AddParams(pkVal)))
		}

		// Field values
		for _, field := range fieldsList {
			if op, exists := opsMap[field]; exists {
				// Поле обновляется
				switch op.Op {
				case activerecord.OpSet:
					values = append(values, fmt.Sprintf("$%d", q.AddParams(op.Value)))
				default:
					return nil, fmt.Errorf("bulk update supports only OpSet, got %d for field %s", op.Op, field)
				}
			} else {
				// Поле не обновляется для этого объекта - используем NULL
				values = append(values, "NULL")
			}
		}

		valueRows = append(valueRows, "    ("+strings.Join(values, ", ")+")")
	}

	q.QueryString += strings.Join(valueRows, ",\n")

	// AS v(pk_fields..., update_fields...)
	q.QueryString += "\n) AS v("

	// Имена колонок в VALUES: pk fields + update fields
	columnNames := make([]string, 0, pkLen+len(fieldsList))
	for _, pkField := range primaryIndex.Fields {
		columnNames = append(columnNames, pkField.Field)
	}
	columnNames = append(columnNames, fieldsList...)

	q.QueryString += strings.Join(columnNames, ", ")
	q.QueryString += ")"

	// WHERE condition для сопоставления по PK
	q.QueryString += "\nWHERE "

	pkConditions := make([]string, 0, len(primaryIndex.Fields))
	for _, pkField := range primaryIndex.Fields {
		pkConditions = append(pkConditions, fmt.Sprintf("t.%s = v.%s", pkField.Field, pkField.Field))
	}
	q.QueryString += strings.Join(pkConditions, " AND ")

	return q, nil
}

func GenerateDelete(tableName string, primaryKey Index, keys [][]any) (*Query, error) {
	if err := primaryKey.validateKeys(keys); err != nil {
		return nil, err
	}

	q := NewDeleteQuery(tableName, primaryKey)

	q.GenerateWhereKeys(primaryKey.MultiField(), keys)

	return q, nil
}

func GenerateInsert(tableName string, pk Index, fieldNames []string, values [][]any, returning []string, conflictAction OnConflictAction, conflictKey Index) (*Query, error) {
	if !conflictKey.Unique {
		return nil, fmt.Errorf("conflict key must be unique")
	}

	bulk := len(values) > 1

	if bulk && conflictAction == IgnoreDuplicate {
		return nil, fmt.Errorf("can't do bulk insert with 'on_conflict_do_nothing' option")
	}

	q := NewInsertQuery(tableName, fieldNames)

	valQ := []string{}

	for _, v := range values {
		if len(v) != len(fieldNames) {
			return nil, fmt.Errorf("fields count not equal to values count")
		}

		innerPlaceholder := make([]string, 0, len(v))
		for _, val := range v {
			if _, ok := val.(DefaultKeyword); ok {
				innerPlaceholder = append(innerPlaceholder, "DEFAULT")
			} else {
				innerPlaceholder = append(innerPlaceholder, fmt.Sprintf("$%d", q.AddParams(val)))
			}
		}

		valQ = append(valQ, "("+strings.Join(innerPlaceholder, ", ")+")")
	}
	q.AddQuery(strings.Join(valQ, ", "))

	switch conflictAction {
	case IgnoreDuplicate:
		q.AddOnConflictDoNothing(fieldNames)
	case Replace:
		q.AddOnConflictDoUpdate(tableName, pk, fieldNames, conflictKey)
	case NoDuplicateAction:
	default:
		return nil, fmt.Errorf("unknown conflict action")
	}

	q.AddReturning(returning)

	return q, nil
}

func (q *Query) ConditionFields(fields []string) {
	if len(fields) > 1 {
		q.AddWhereQuery("(" + strings.Join(fields, ", ") + ")")
		return
	}

	q.AddWhereQuery(fields[0])
}
