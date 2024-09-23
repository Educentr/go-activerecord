package generator

import (
	"strings"
	"text/template"

	"github.com/mailru/activerecord/pkg/iproto/util/text"
)

var BaseTemplateFuncs = template.FuncMap{
	"split":      strings.Split,
	"trimPrefix": strings.TrimPrefix,
	"hasPrefix":  strings.HasPrefix,
	"snakeCase":  text.ToSnakeCase,
}
