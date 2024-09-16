package postgres

import (
	"fmt"
	"strings"

	"github.com/mailru/activerecord/pkg/activerecord"
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

func NewQuery(tableName string, fieldNames []string, i Index) *Query {
	return &Query{
		QueryString: fmt.Sprintf(`SELECT %s FROM %s WHERE %s%s`,
			strings.Join(fieldNames, ", "),
			QuoteIdentifier(tableName),
			i.ConditionString(),
			i.ConditionFieldsString(),
		),
	}
}

func (q *Query) AddWhereParams(key any) int {
	q.Params = append(q.Params, key)

	return len(q.Params)
}

func (q *Query) AddWhereCondition(cond string, key []any) {
	q.Params = append(q.Params, key...)
	q.QueryString += cond
}

func (q *Query) AddQuery(cond string) {
	q.QueryString += cond
}

func (q *Query) AddLimitOffset(limit, offset uint16) {
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

func GenerateSelect(tableName string, fieldNames []string, index Index, keys [][]any, offset, limit uint16, cursor CursorPosition) (*Query, error) {
	if err := index.validateKeys(keys); err != nil {
		return nil, err
	}

	if err := index.ValidateCursor(cursor); err != nil {
		return nil, err
	}

	bulkSelect := len(keys) > 1
	oneRowResult := limit == 1 || (!bulkSelect && index.Unique)

	q := NewQuery(tableName, fieldNames, index)

	// ToDo work with array fields
	// $pg_request->{query} .= ' && $' . $i++;
	// push @{$pg_request->{params}}, '{' . MR::Pg->encode_array_field($request->{keys}) . '}';

	if bulkSelect {
		placeholders := make([]string, 0, len(keys))
		if index.MultiField() {
			for _, key := range keys {
				innerPlaceholder := make([]string, 0, len(key))

				for _, kField := range key {
					innerPlaceholder = append(innerPlaceholder, fmt.Sprintf("$%d", q.AddWhereParams(kField)))
				}

				placeholders = append(placeholders, "("+strings.Join(innerPlaceholder, ", ")+")")
			}
		} else {
			for _, key := range keys {
				placeholders = append(placeholders, fmt.Sprintf("$%d", q.AddWhereParams(key[0])))
			}
		}

		q.QueryString += " IN (" + strings.Join(placeholders, ", ") + ")"
	} else {
		if index.MultiField() {
			innerPlaceholder := make([]string, 0, len(keys[0]))

			for _, kField := range keys[0] {
				innerPlaceholder = append(innerPlaceholder, fmt.Sprintf("$%d", q.AddWhereParams(kField)))
			}

			q.QueryString += " = (" + strings.Join(innerPlaceholder, ", ") + ")"
		} else {
			q.QueryString += fmt.Sprintf(" = $%d", q.AddWhereParams(keys[0][0]))
		}
	}

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
