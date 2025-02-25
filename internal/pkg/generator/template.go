package generator

import (
	"strings"
	"text/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/textutil"
)

var BaseTemplateFuncs = template.FuncMap{
	"split":      strings.Split,
	"trimPrefix": strings.TrimPrefix,
	"hasPrefix":  strings.HasPrefix,
	"snakeCase":  textutil.ToSnakeCase,
}
