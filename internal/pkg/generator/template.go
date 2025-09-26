package generator

import (
	"strings"
	"text/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/textutil"
)

var stringTemplateFuncs = template.FuncMap{
	"split":      strings.Split,
	"trimPrefix": strings.TrimPrefix,
	"hasPrefix":  strings.HasPrefix,
	"snakeCase":  textutil.ToSnakeCase,
}

var mathTemplateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
}
