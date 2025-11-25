package postgres

import (
	"errors"
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
	// Поддерживаемые операторы
	supportedOperators := map[string]bool{
		"=":          true,
		">":          true,
		"<":          true,
		">=":         true,
		"<=":         true,
		"!=":         true,
		"is null":    true,
		"is not null": true,
	}

	for _, ind := range cl.Indexes {
		for fldNum, cond := range ind.Conditions {
			fld := cl.Fields[fldNum]
			format, err := GetFormat(fld.Format)
			if err != nil {
				return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckInternalError}
			}

			// Проверка поддержки оператора
			if !supportedOperators[cond.ConditionType] {
				return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionOperatorUnsupported}
			}

			// Проверка значений для IS NULL/IS NOT NULL
			if cond.IsNullCheck {
				if len(cond.Value) > 0 {
					return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionNullCheckWithValues}
				}
				continue // NULL проверки не требуют дальнейшей валидации
			}

			// Проверка наличия значений для не-NULL операторов
			if len(cond.Value) == 0 {
				return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionHasNotValue}
			}

			// Проверка для операторов равенства/неравенства с пустой строкой
			if (cond.ConditionType == "=" || cond.ConditionType == "!=") && len(cond.Value) == 1 && cond.Value[0] == "" {
				// Пустая строка допустима только для строковых типов
				isStringType := fld.Format == "string"
				if !isStringType {
					return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: fmt.Errorf("empty string condition is only valid for string fields")}
				}
			}

			// Проверка количества значений для операторов сравнения
			if cond.ConditionType == ">" || cond.ConditionType == "<" || cond.ConditionType == ">=" || cond.ConditionType == "<=" {
				if len(cond.Value) != 1 {
					return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionValuesMismatch}
				}

				// Проверка совместимости типа поля с оператором сравнения
				// Операторы сравнения требуют числовых или временных типов
				isCompatible := false
				for _, numType := range NumericFormatT {
					if fld.Format == numType.TypeName {
						isCompatible = true
						break
					}
				}
				if !isCompatible {
					for _, floatType := range FloatFormatT {
						if fld.Format == floatType.TypeName {
							isCompatible = true
							break
						}
					}
				}
				if !isCompatible {
					for _, dateType := range DateFormatT {
						if fld.Format == dateType.TypeName {
							isCompatible = true
							break
						}
					}
				}
				if !isCompatible {
					return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckIndexConditionTypeIncompatible}
				}
			}

			// Проверка что десериализатор из строки будет работать корректно
			deserializers := format.StringDeserializer()
			if len(deserializers) == 0 && len(cond.Value) > 0 {
				// Если нет десериализатора, значения будут использоваться как строки напрямую
				// Это нормально для строковых типов
			}
		}
	}

	return nil
}

func (b BackendGenerator) CheckFlags(cl *ds.RecordPackage) error {
	for fieldName, flagDecl := range cl.FlagMap {
		fieldNum, ok := cl.FieldsMap[fieldName]
		if !ok {
			continue
		}

		field := cl.Fields[fieldNum]

		// PostgreSQL поддерживает ТОЛЬКО signed типы
		var bitCapacity int
		switch field.Format {
		case "int8":
			bitCapacity = 8
		case "int16":
			bitCapacity = 16
		case "int32":
			bitCapacity = 32
		case "int64":
			bitCapacity = 64
		case "uint8", "uint16", "uint32", "uint64":
			return &arerror.ErrCheckPackageFieldDecl{
				Pkg:   cl.Namespace.PackageName,
				Field: fieldName,
				Err:   fmt.Errorf("PostgreSQL не поддерживает unsigned типы. Используйте signed типы (int8, int16, int32, int64) для флагов. Тип поля: %s", field.Format),
			}
		default:
			return &arerror.ErrCheckPackageFieldDecl{
				Pkg:   cl.Namespace.PackageName,
				Field: fieldName,
				Err:   errors.New("флаги могут быть объявлены только на целочисленных типах (int8, int16, int32, int64)"),
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
