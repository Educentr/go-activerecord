package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

// Парсинг флагов. В описании модели можно указать, что целочисленное значение используется для хранения
// битовых флагов. В этом случае на поле навешиваются мутаторы SetFlag и ClearFlag
func ParseFlags(dst *ds.RecordPackage, fields []*ast.Field) error {
	for _, field := range fields {
		if field.Names == nil || len(field.Names) != 1 {
			return &arerror.ErrParseFlagDecl{Err: arerror.ErrNameDeclaration}
		}

		newflag := ds.FlagDeclaration{
			Name:  field.Names[0].Name,
			Flags: []ds.FlagItem{},
		}

		tagParam, err := splitTag(field, CheckFlagEmpty, map[TagNameType]ParamValueRule{})
		if err != nil {
			return &arerror.ErrParseFlagDecl{Name: newflag.Name, Err: err}
		}

		for _, kv := range tagParam {
			switch kv[0] {
			case "flags":
				rawFlags := strings.Split(kv[1], ",")
				newflag.BitCount = len(rawFlags)
				newflag.Flags = make([]ds.FlagItem, 0, len(rawFlags))

				// Проверка на дубликаты имен
				seenNames := make(map[string]bool)

				for i, flag := range rawFlags {
					trimmed := strings.TrimSpace(flag)
					if trimmed == "" || trimmed == "_" {
						// Пропускаем пустые и underscore флаги
						continue
					}

					// Проверка на дубликат
					if seenNames[trimmed] {
						return &arerror.ErrParseFlagDecl{
							Name: newflag.Name,
							Err:  fmt.Errorf("дублирующееся имя флага: %s", trimmed),
						}
					}
					seenNames[trimmed] = true

					newflag.Flags = append(newflag.Flags, ds.FlagItem{
						Name:     trimmed,
						Position: i,
					})
				}
			default:
				return &arerror.ErrParseFlagTagDecl{Name: newflag.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagUnknown}
			}
		}

		fldNum, ok := dst.FieldsMap[newflag.Name]
		if !ok {
			return &arerror.ErrParseFlagDecl{Name: newflag.Name, Err: arerror.ErrFieldNotExist}
		}

		foundSet, foundClear := false, false

		for _, mut := range dst.Fields[fldNum].Mutators {
			if mut == ds.SetBitMutator {
				foundSet = true
			}

			// Bug fix: проверяем ClearBitMutator вместо SetBitMutator
			if mut == ds.ClearBitMutator {
				foundClear = true
			}
		}

		if !foundSet {
			dst.Fields[fldNum].Mutators = append(dst.Fields[fldNum].Mutators, ds.SetBitMutator)
		}

		if !foundClear {
			dst.Fields[fldNum].Mutators = append(dst.Fields[fldNum].Mutators, ds.ClearBitMutator)
		}

		if err = dst.AddFlag(newflag); err != nil {
			return err
		}
	}

	return nil
}
