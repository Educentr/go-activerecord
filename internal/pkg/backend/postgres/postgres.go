package postgres

import (
	"embed"
	"log"
	"text/template"

	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	postgresPkg "github.com/Educentr/go-activerecord/pkg/postgres"
)

type BackendGenerator struct {
	availFormat map[ds.Format]struct{}
}

//go:embed tmpl/pkg/*
var Templates embed.FS

func (b *BackendGenerator) Init() {
	b.availFormat = make(map[ds.Format]struct{}, len(AllFormatT))

	for _, form := range AllFormatT {
		b.availFormat[form.TypeName] = struct{}{}
	}
}

func (b BackendGenerator) Name() ds.Backend {
	return "postgres"
}

func (b BackendGenerator) Aliases() []ds.Backend {
	return []ds.Backend{}
}

func (b BackendGenerator) Templates() embed.FS {
	return Templates
}

func (b BackendGenerator) TemplatePath() string {
	return "tmpl/pkg"
}

func (b BackendGenerator) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// ToDo merge with octopus
		"packerParam": func(format ds.Format) ds.FormatParam {
			f, err := GetFormat(format)
			if err != nil {
				log.Fatal(err)
			}

			return f
		},
		"indexOrder": func(iField ds.IndexField) postgresPkg.Order {
			switch iField.Order {
			case ds.IndexOrderAsc:
				return postgresPkg.ASC
			case ds.IndexOrderDesc:
				return postgresPkg.DESC
			default:
				log.Fatal("invalid index field order")
			}

			return postgresPkg.ASC
		},
	}
}
