package postgres

// conditionListOperators сопоставляет оператор условия индекса с SQL-оператором
// для списка значений. Таблица одна на генератор DDL и рантайм: предикат
// частичного индекса и условие запроса по этому индексу обязаны совпадать,
// иначе планировщик не применит индекс к запросу, а у уникального индекса
// меняется смысл ограничения
var conditionListOperators = map[string]string{
	"=":  "IN",
	"!=": "NOT IN",
}

// ConditionListOperator возвращает SQL-оператор, которым условие индекса
// сравнивается со списком значений: `=` → `IN`, `!=` → `NOT IN`.
//
// Для остальных операторов возвращается ok=false: сравнение (`>`, `<`, `>=`,
// `<=`) со списком значений смысла не имеет и отвергается checker-ом argen.
func ConditionListOperator(operator string) (string, bool) {
	op, ex := conditionListOperators[operator]
	return op, ex
}
