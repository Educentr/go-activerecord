package checker

import (
	"errors"
	"log"

	"github.com/mailru/activerecord/internal/pkg/arerror"
	"github.com/mailru/activerecord/internal/pkg/ds"
)

// Checker структура описывающая checker
type Checker struct {
	files map[string]*ds.RecordPackage
}

// Init конструктор checker-а
func Init(files map[string]*ds.RecordPackage) *Checker {
	checker := Checker{
		files: files,
	}

	return &checker
}

// checkBackend проверка на указание бекенда
// В данный момент поддерживается один и толлько один бекенд
func checkBackend(cl *ds.RecordPackage) error {
	if len(cl.Backends) == 0 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckBackendEmpty}
	}

	if len(cl.Backends) > 1 {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckPkgBackendToMatch}
	}

	return nil
}

// checkLinkedObject проверка существования сущностей на которые ссылаются другие сущности
func checkLinkedObject(cl *ds.RecordPackage, linkedObjects map[string]string) error {
	for _, fobj := range cl.FieldsObjectMap {
		if _, ok := linkedObjects[fobj.ObjectName]; !ok {
			return &arerror.ErrCheckPackageLinkedDecl{Pkg: cl.Namespace.PackageName, Object: fobj.ObjectName, Err: arerror.ErrCheckObjectNotFound}
		}
	}

	return nil
}

// checkNamespace проверка правильного описания неймспейса у сущности
func checkNamespace(ns ds.NamespaceDeclaration) error {
	if ns.PackageName == "" || ns.PublicName == "" {
		return &arerror.ErrCheckPackageNamespaceDecl{Pkg: ns.PackageName, Name: ns.PublicName, Err: arerror.ErrCheckEmptyNamespace}
	}

	return nil
}

// checkFields функция проверки правильности описания полей структуры
// - указан допустимый тип полей
// - описаны все необходимые сериализаторы для полей с сериализацией
// - поля с мутаторами не могут быть праймари ключом
// - поля с мутаторами не могут быть сериализованными
// - поля с мутаторами не могут являться ссылками на другие сущности
// - сериализуемые поля не могут быть ссылками на другие сущности
// - есть первичный ключ
// - имена сущностей на которые ссылаемся не могут пересекаться с именами полей
//
//nolint:gocognit,gocyclo
func checkFields(cl *ds.RecordPackage) error {
	primaryFound := false

	for _, ind := range cl.Indexes {
		if len(ind.Fields) == 0 {
			return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: ind.Name, Err: arerror.ErrCheckFieldIndexEmpty}
		}
	}

	for _, fld := range cl.Fields {
		if (fld.Format == "string" || fld.Format == "[]byte") && fld.Size == 0 {
			log.Printf("Warn: field `%s` declaration. Field with type string or []byte not contain size.", fld.Name)
		}

		if len(fld.Serializer) > 0 {
			if _, ex := cl.SerializerMap[fld.Serializer[0]]; len(cl.SerializerMap) == 0 || !ex {
				return &arerror.ErrCheckPackageFieldDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerNotFound}
			}
		}

		customMutCnt := 0
		if len(fld.Mutators) > 0 {
			fieldMutatorsChecker := ds.GetFieldMutatorsChecker()

			for _, m := range fld.Mutators {
				_, ex := fieldMutatorsChecker[m]

				md, ok := cl.MutatorMap[m]
				if ok {
					customMutCnt++
					if customMutCnt > 1 {
						return &arerror.ErrCheckPackageFieldMutatorDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Mutator: m, Err: arerror.ErrParseFieldMutatorInvalid}
					}
				}

				if !ok && !ex {
					return &arerror.ErrCheckPackageFieldMutatorDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Mutator: m, Err: arerror.ErrParseFieldMutatorInvalid}
				}

				if len(md.PartialFields) > 0 && len(fld.Serializer) == 0 {
					return &arerror.ErrCheckPackageFieldMutatorDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Mutator: m, Err: arerror.ErrParseFieldMutatorTypeHasNotSerializer}
				}
			}

			if fld.PrimaryKey {
				return &arerror.ErrCheckPackageFieldMutatorDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Mutator: string(fld.Mutators[0]), Err: arerror.ErrCheckFieldMutatorConflictPK}
			}

			if fld.ObjectLink != "" {
				return &arerror.ErrCheckPackageFieldMutatorDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Mutator: string(fld.Mutators[0]), Err: arerror.ErrCheckFieldMutatorConflictObject}
			}
		}

		if len(fld.Serializer) > 0 && fld.ObjectLink != "" {
			return &arerror.ErrCheckPackageFieldMutatorDecl{Pkg: cl.Namespace.PackageName, Field: fld.Name, Err: arerror.ErrCheckFieldSerializerConflictObject}
		}

		if fo, ex := cl.FieldsObjectMap[fld.Name]; ex {
			return &arerror.ErrParseTypeFieldStructDecl{Name: fo.Name, Err: arerror.ErrRedefined}
		}

		if fld.PrimaryKey {
			primaryFound = true
		}
	}

	if len(cl.Fields) > 0 && !primaryFound {
		return &arerror.ErrCheckPackageIndexDecl{Pkg: cl.Namespace.PackageName, Index: "primary", Err: arerror.ErrIndexNotExist}
	}

	return nil
}

// Check основная функция, которая запускает процесс проверки
// Должна вызываться только после окончания процесса парсинга всех деклараций
func Check(files map[string]*ds.RecordPackage, linkedObjects map[string]string) error {
	for _, cl := range files {
		if err := checkBackend(cl); err != nil {
			return err
		}

		if err := checkNamespace(cl.Namespace); err != nil {
			return err
		}

		if err := checkServerConfig(cl); err != nil {
			return err
		}

		if err := checkLinkedObject(cl, linkedObjects); err != nil {
			return err
		}

		backendChecker, err := getBackendSpecificChecker(cl.Backends[0])
		if err != nil {
			if errors.Is(err, ErrBackendNotImplemented) {
				return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Backend: cl.Backends[0], Err: arerror.ErrGeneratorBackendNotImplemented}
			}
		}

		if err := backendChecker.check(cl); err != nil {
			return err
		}

		if err := backendChecker.checkFields(cl); err != nil {
			return err
		}

		if err := backendChecker.checkNamespace(cl); err != nil {
			return err
		}

		if err := checkFields(cl); err != nil {
			return err
		}
	}

	return nil
}

func checkServerConfig(cl *ds.RecordPackage) error {
	if cl.ServerConfKey == "" {
		return &arerror.ErrCheckPackageDecl{Pkg: cl.Namespace.PackageName, Err: arerror.ErrCheckServerEmpty}
	}

	return nil
}
