package checker

import (
	"strconv"

	"github.com/mailru/activerecord/internal/pkg/arerror"
	"github.com/mailru/activerecord/internal/pkg/ds"
	"github.com/mailru/activerecord/pkg/activerecord"
	"github.com/mailru/activerecord/pkg/octopus"
)

type octopusChecker struct {
	availFormat     map[activerecord.Format]struct{}
	procAvailFormat map[activerecord.Format]struct{}
}

func CreateOctopusChecker() backendChecker {
	checker := &octopusChecker{
		availFormat:     make(map[activerecord.Format]struct{}, len(octopus.AllFormat)),
		procAvailFormat: make(map[activerecord.Format]struct{}, len(octopus.AllProcFormat)),
	}

	for _, form := range octopus.AllFormat {
		checker.availFormat[form] = struct{}{}
	}

	for _, form := range octopus.AllProcFormat {
		checker.procAvailFormat[form] = struct{}{}
	}

	return checker
}

func (c *octopusChecker) check(cl *ds.RecordPackage) error {
	if len(cl.Fields) > 0 {
		_, err := strconv.ParseInt(cl.Namespace.ObjectName, 10, 64)
		if err != nil {
			return &arerror.ErrCheckPackageNamespaceDecl{Pkg: cl.Namespace.PackageName, Name: cl.Namespace.ObjectName, Err: arerror.ErrCheckFieldInvalidFormat}
		}
	}

	return nil
}

func (c *octopusChecker) checkFields(cl *ds.RecordPackage) error {
	if len(cl.Fields) > 0 && len(cl.ProcOutFields) > 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsManyDecl}
	}

	for _, fld := range cl.Fields {
		if _, ex := c.availFormat[fld.Format]; !ex {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldInvalidFormat}
		}
	}

	if len(cl.Fields) == 0 && len(cl.ProcOutFields) == 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsEmpty}
	}

	return c.checkProcFields(cl)
}

func (c *octopusChecker) checkProcFields(cl *ds.RecordPackage) error {
	if !cl.ProcOutFields.Validate() {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsOrderDecl}
	}

	for _, fld := range cl.ProcOutFields.List() {
		if _, ex := c.availFormat[fld.Format]; !ex {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldInvalidFormat}
		}

		if len(fld.Serializer) > 0 {
			if _, ex := cl.SerializerMap[fld.Serializer[0]]; !ex {
				return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerNotFound}
			}
		}

		if fld.Type == 0 {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldTypeNotFound}
		}
	}

	for _, fld := range cl.ProcInFields {
		if _, ex := c.procAvailFormat[fld.Format]; !ex {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldInvalidFormat}
		}

		if len(fld.Serializer) > 0 {
			if _, ex := cl.SerializerMap[fld.Serializer[0]]; !ex {
				return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerNotFound}
			}
		}

		if fld.Format != octopus.String && len(fld.Serializer) == 0 {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerNotFound}
		}

		if fld.Type == 0 {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldTypeNotFound}
		}
	}

	return nil
}
