package postgres

import (
	"fmt"
	"strings"

	"github.com/Educentr/go-activerecord/pkg/activerecord"
)

//type WhereCondition

const (
	Backend activerecord.Backend = "postgres"
)

type QueryBuilderState uint8

const (
	QueryBuilderStateWhere QueryBuilderState = iota
	QueryBuilderStateOrderBy
	QueryBuilderStateLimit
	QueryBuilderStateOffset
)

type Query struct {
	QueryString string
	Params      []any
	// state       QueryBuilderState
}

func NewSelectQuery(tableName string, fieldNames []string, i Index) *Query {
	// ToDo quote field names
	q := &Query{
		QueryString: fmt.Sprintf(`SELECT %s FROM %s`,
			strings.Join(fieldNames, ", "),
			QuoteIdentifier(tableName),
		),
	}

	q.AddWhereBlock(i.Conditions(), i.ConditionFields())

	return q
}

func NewUpdateQuery(tableName string) *Query {
	return &Query{
		QueryString: fmt.Sprintf(`UPDATE %s SET `,
			QuoteIdentifier(tableName),
		),
	}
}

func NewDeleteQuery(tableName string, pk Index) *Query {
	q := &Query{
		QueryString: fmt.Sprintf(`DELETE FROM %s`,
			QuoteIdentifier(tableName),
		),
	}

	q.AddWhereBlock(pk.ConditionFields())

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

func (q *Query) AddNoConflictDoNothing(fieldNames []string) {
	q.QueryString += "ON CONFLICT DO NOTHING"
}

func (q *Query) AddNoConflictDoUpdate(tableName string, pk Index, fieldNames []string) {
	pkfields := make(map[string]struct{}, len(pk.Fields))
	updateFields := []string{}

	for _, pkf := range pk.Fields.GetFieldNames() {
		updateFields = append(updateFields, fmt.Sprintf("%s=%s.%s", pkf, tableName, pkf))
		pkfields[pkf] = struct{}{}
	}

	for _, f := range fieldNames {
		if _, ex := pkfields[f]; ex {
			continue
		}

		updateFields = append(updateFields, fmt.Sprintf("%s=EXCLUDED.%s", f, f))
	}

	q.QueryString += fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(pk.Fields.GetFieldNames(), ", "),
		strings.Join(updateFields, ", "),
	)
}

func (q *Query) AddParams(key ...any) int {
	q.Params = append(q.Params, key...)

	return len(q.Params)
}

func (q *Query) AddQuery(cond string) {
	q.QueryString += cond + " "
}

func (q *Query) AddWhereCondition(cond string, key []any) {
	q.AddParams(key...)
	q.AddQuery(cond)
}

func (q *Query) AddSetStatement(field string, key any) {
	q.AddQuery(field + " = $" + fmt.Sprintf("%d", key))
}

func (q *Query) AddWhereBlock(cond ...string) {
	q.QueryString += " WHERE "
	for _, c := range cond {
		q.AddQuery(c)
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

func GenerateSelectAll(tableName string, fieldNames []string) (string, error) {
	return fmt.Sprintf("SELECT %s FROM %s LIMIT %d", strings.Join(fieldNames, ", "), QuoteIdentifier(tableName), MaxLimit), nil
}

func GenerateSelect(tableName string, fieldNames []string, index Index, keys [][]any, offset, limit uint32, cursor CursorPosition) (*Query, error) {
	if err := index.validateKeys(keys); err != nil {
		return nil, err
	}

	if err := index.ValidateCursor(cursor); err != nil {
		return nil, err
	}

	bulkSelect := len(keys) > 1
	oneRowResult := limit == 1 || (!bulkSelect && index.Unique)

	q := NewSelectQuery(tableName, fieldNames, index)

	// ToDo work with array fields
	// $pg_request->{query} .= ' && $' . $i++;
	// push @{$pg_request->{params}}, '{' . MR::Pg->encode_array_field($request->{keys}) . '}';

	index.GenerateWhereKeys(q, keys)

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

	// 	if($opts{order}) {
	// 		my $order = _prepare_order($opts{order}, $fields_list);
	// 		$pg_request->{query} .= ' ORDER BY '. join(', ', map { $field_sql_deserialized_name{$_->{field}}.' '.$_->{order} } @$order);
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

func GenerateUpdate(tableName string, primaryIndex Index, updates []UpdateParams) (*Query, error) {
	// ToDo generate bulk update
	isBulk := len(updates) > 1
	if isBulk {
		return nil, fmt.Errorf("bulk update not implemented")
	}

	q := NewUpdateQuery(tableName)

	for num, u := range updates {
		if len(u.PK) != len(primaryIndex.Fields) {
			return nil, fmt.Errorf("primary key length (%+v) not equal to index fields in update %d", u.PK, num)
		}

		for _, op := range u.Ops {
			// ToDo serializers
			operation := op.Field + " ="
			returning := []string{}

			switch op.Op {
			case activerecord.OpSet:
				operation += " $" + fmt.Sprintf("%d", q.AddParams(op.Value))
			case activerecord.OpAdd:
				operation += op.Field + " + $" + fmt.Sprintf("%d", q.AddParams(op.Value))
				returning = append(returning, op.Field)
			case activerecord.OpAnd:
				operation += op.Field + " & $" + fmt.Sprintf("%d", q.AddParams(op.Value))
				returning = append(returning, op.Field)
			default:
				return nil, fmt.Errorf("unknown operation %d or not implemented", op.Op)
			}

			q.AddQuery(operation)

			q.AddWhereBlock(primaryIndex.ConditionFields())

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
	}

	return q, nil
}

func GenerateDelete(tableName string, primaryKey Index, keys [][]any) (*Query, error) {
	if err := primaryKey.validateKeys(keys); err != nil {
		return nil, err
	}

	q := NewDeleteQuery(tableName, primaryKey)
	primaryKey.GenerateWhereKeys(q, keys)

	return q, nil
}

func GenerateInsert(tableName string, pk Index, fieldNames []string, values [][]any, returning []string, conflictAction OnConflictAction) (*Query, error) {
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
		q.AddNoConflictDoNothing(fieldNames)
	case UpdateDuplicate:
		q.AddNoConflictDoUpdate(tableName, pk, fieldNames)
	case NoDuplicateAction:
	default:
		return nil, fmt.Errorf("unknown conflict action")
	}

	q.AddReturning(returning)

	return q, nil
}
