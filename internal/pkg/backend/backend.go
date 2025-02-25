package backend

import (
	"embed"
	"errors"
	"html/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/backend/octopus"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/backend/postgres"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

type BackendDriver interface {
	Init()
	Check(cl *ds.RecordPackage) error
	CheckFields(cl *ds.RecordPackage) error
	CheckIndexes(cl *ds.RecordPackage) error
	CheckNamespace(cl *ds.RecordPackage) error
	Name() ds.Backend
	Aliases() []ds.Backend
	Templates() embed.FS
	TemplatePath() string
	TemplateFuncs() template.FuncMap
}

var registeredBackend = map[ds.Backend]BackendDriver{}

var (
	ErrBackendNotImplemented = errors.New("backend not implemented")
)

// ToDo Tarantool16 / Tarantool2

func RegisterBackend() {
	// ToDo сделать проверку на то, что имена бекендов не пересекаются
	for _, b := range []BackendDriver{&postgres.BackendGenerator{}, &octopus.BackendGenerator{}} {
		b.Init()

		registeredBackend[b.Name()] = b

		for _, bn := range b.Aliases() {
			registeredBackend[bn] = b
		}
	}
}

func GetBackendByName(name ds.Backend) (BackendDriver, error) {
	b, ex := registeredBackend[name]
	if !ex {
		return nil, ErrBackendNotImplemented
	}

	return b, nil
}
