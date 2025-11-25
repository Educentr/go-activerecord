package generator

import (
	"strings"
	"text/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/textutil"
)

var stringTemplateFuncs = template.FuncMap{
	"split":         strings.Split,
	"trimPrefix":    strings.TrimPrefix,
	"hasPrefix":     strings.HasPrefix,
	"snakeCase":     textutil.ToSnakeCase,
	"operatorToSQL": operatorToSQL,
}

// operatorToSQL преобразует оператор в SQL формат
func operatorToSQL(op string) string {
	switch strings.ToLower(op) {
	case "is null":
		return "IS NULL"
	case "is not null":
		return "IS NOT NULL"
	default:
		return strings.ToUpper(op)
	}
}

var mathTemplateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
}
