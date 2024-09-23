package generator

import (
	"embed"
	"log"
	"text/template"

	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/pkg/postgres"
)

//go:embed tmpl/postgres/pkg/*
var postgresTemplatesPath embed.FS

var tmplPostgresPath = "tmpl/postgres/pkg"

var PostgresTemplateFuncs = template.FuncMap{
	"indexOrder": func(iField ds.IndexField) postgres.Order {
		switch iField.Order {
		case ds.IndexOrderAsc:
			return postgres.ASC
		case ds.IndexOrderDesc:
			return postgres.DESC
		default:
			log.Fatal("invalid index field order")
		}

		return postgres.ASC
	},
}
