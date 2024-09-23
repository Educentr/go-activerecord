package checker

import (
	"errors"

	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/pkg/activerecord"
	"github.com/Educentr/go-activerecord/pkg/octopus"
	"github.com/Educentr/go-activerecord/pkg/postgres"
)

type backendChecker interface {
	check(cl *ds.RecordPackage) error
	checkFields(cl *ds.RecordPackage) error
	checkNamespace(cl *ds.RecordPackage) error
}

var (
	ErrBackendNotImplemented = errors.New("backend not implemented")
)

func getBackendSpecificChecker(backend activerecord.Backend) (checker backendChecker, err error) {
	switch backend {
	case octopus.BackendTarantool:
		fallthrough
	case octopus.Backend:
		checker = CreateOctopusChecker()
	case postgres.Backend:
		checker = CreatePostgresChecker()
	default:
		err = ErrBackendNotImplemented
	}

	return
}
