package checker

import (
	"regexp"

	"github.com/Educentr/go-activerecord/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/pkg/activerecord"
	"github.com/Educentr/go-activerecord/pkg/postgres"
)

type postgresChecker struct {
	availFormat map[activerecord.Format]struct{}
}

func CreatePostgresChecker() backendChecker {
	checker := &postgresChecker{
		availFormat: make(map[activerecord.Format]struct{}, len(postgres.AllFormat)),
	}

	for _, form := range postgres.AllFormat {
		checker.availFormat[form] = struct{}{}
	}

	return checker
}

func (c *postgresChecker) check(cl *ds.RecordPackage) error {
	return nil
}

func (c *postgresChecker) checkFields(cl *ds.RecordPackage) error {
	if len(cl.ProcOutFields) > 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsProcNotImpl}
	}

	for _, fld := range cl.Fields {
		if _, ex := c.availFormat[fld.Format]; !ex {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldInvalidFormat}
		}
	}

	if len(cl.Fields) == 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsEmpty}
	}

	return nil
}

var rxCanonicalTableName = regexp.MustCompile("^[a-z0-9]*$")

func (c *postgresChecker) checkNamespace(cl *ds.RecordPackage) error {
	if !rxCanonicalTableName.MatchString(cl.Namespace.ObjectName) {
		return &arerror.ErrCheckPackageNamespaceDecl{Pkg: cl.Namespace.PackageName, Name: cl.Namespace.ObjectName, Err: arerror.ErrTableNameNotCanonical}

	}
	return nil
}
