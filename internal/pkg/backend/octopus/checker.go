package octopus

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

func (c BackendGenerator) Check(cl *ds.RecordPackage) error {
	if len(cl.Fields) > 0 {
		_, err := strconv.ParseInt(cl.Namespace.ObjectName, 10, 64)
		if err != nil {
			return &arerror.ErrCheckPackageNamespaceDecl{Pkg: cl.Namespace.PackageName, Name: cl.Namespace.ObjectName, Err: arerror.ErrCheckFieldInvalidFormat}
		}
	}

	if cl.EnableSelectAll {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckSelectAllNotSupported}
	}

	return nil
}

func (c BackendGenerator) CheckFields(cl *ds.RecordPackage) error {
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

	return c.CheckProcFields(cl)
}

func (c BackendGenerator) CheckProcFields(cl *ds.RecordPackage) error {
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
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldInvalidProcFormat}
		}

		if len(fld.Serializer) > 0 {
			if _, ex := cl.SerializerMap[fld.Serializer[0]]; !ex {
				return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerNotFound}
			}
		}

		if fld.Format != "string" && len(fld.Serializer) == 0 {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerNotFound}
		}

		if fld.Type == 0 {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldTypeNotFound}
		}
	}

	return nil
}

func (c BackendGenerator) CheckNamespace(_ *ds.RecordPackage) error {
	return nil
}

func (c BackendGenerator) CheckIndexes(cl *ds.RecordPackage) error {
	for _, ind := range cl.Indexes {
		if len(ind.Conditions) != 0 {
			return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionNotSupported}
		}

		if ind.SelectorCount != "" {
			return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexCountNotSupported}
		}
	}

	return nil
}

func (c BackendGenerator) CheckFlags(cl *ds.RecordPackage) error {
	for fieldName, flagDecl := range cl.FlagMap {
		fieldNum, ok := cl.FieldsMap[fieldName]
		if !ok {
			continue
		}

		field := cl.Fields[fieldNum]

		// Octopus поддерживает uint и int типы
		var bitCapacity int
		switch field.Format {
		case "int8", "uint8":
			bitCapacity = 8
		case "int16", "uint16":
			bitCapacity = 16
		case "int32", "uint32":
			bitCapacity = 32
		case "int64", "uint64":
			bitCapacity = 64
		default:
			return &arerror.ErrCheckPackageFieldDecl{
				Pkg:   cl.Namespace.PackageName,
				Field: fieldName,
				Err:   errors.New("флаги могут быть объявлены только на целочисленных типах"),
			}
		}

		if flagDecl.BitCount > bitCapacity {
			return &arerror.ErrCheckPackageFieldDecl{
				Pkg:   cl.Namespace.PackageName,
				Field: fieldName,
				Err:   fmt.Errorf("количество флагов (%d) превышает битовую емкость (%d) для типа %s", flagDecl.BitCount, bitCapacity, field.Format),
			}
		}
	}

	return nil
}
