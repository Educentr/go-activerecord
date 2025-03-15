package postgres

import (
	"fmt"
	"regexp"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

var rxCanonicalTableName = regexp.MustCompile("^[a-z0-9_]*$")

func (b BackendGenerator) Check(cl *ds.RecordPackage) error {
	return nil
}

func (b BackendGenerator) CheckFields(cl *ds.RecordPackage) error {
	if len(cl.ProcOutFields) > 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsProcNotImpl}
	}

	for _, fld := range cl.Fields {
		if _, ex := b.availFormat[fld.Format]; !ex {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldInvalidFormat}
		}

		f, err := GetFormat(fld.Format)
		if err != nil {
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: fmt.Errorf("can't get format %s", err)}
		}

		if fld.PrimaryKey && f.PackFunc() != "" {
			// ToDo имплементировать возможность использования типов с сериализацией на уровне БД как первичный ключ или часть первичного ключа
			return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: fmt.Errorf("type with dbserializer not supported as primary key")}
		}
	}

	if len(cl.Fields) == 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckFieldsEmpty}
	}

	return nil
}

func (b BackendGenerator) CheckNamespace(cl *ds.RecordPackage) error {
	if !rxCanonicalTableName.MatchString(cl.Namespace.ObjectName) {
		return &arerror.ErrCheckPackageNamespaceDecl{Pkg: cl.Namespace.PackageName, Name: cl.Namespace.ObjectName, Err: arerror.ErrTableNameNotCanonical}
	}
	return nil
}

func (b BackendGenerator) CheckIndexes(cl *ds.RecordPackage) error {
	for _, ind := range cl.Indexes {
		for fldNum, cond := range ind.Conditions {
			fld := cl.Fields[fldNum]
			_, err := GetFormat(fld.Format)
			if err != nil {
				return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckInternalError}
			}

			if cond.ConditionType != "=" {
				return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionNotSupported}
			}

			if len(cond.Value) == 0 {
				return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionHasNotValue}
			}

			// ToDo проверка на то, что десериализатор из строки будет работать корректно и сгенерируется валидный код
		}
	}

	return nil
}
