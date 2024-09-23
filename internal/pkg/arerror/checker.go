package arerror

import (
	"errors"

	"github.com/Educentr/go-activerecord/pkg/activerecord"
)

var (
	ErrCheckBackendEmpty                   = errors.New("backend empty")
	ErrCheckBackendUnknown                 = errors.New("backend unknown")
	ErrCheckEmptyNamespace                 = errors.New("empty namespace")
	ErrCheckPkgBackendToMatch              = errors.New("many backends for one class not supported yet")
	ErrCheckFieldSerializerNotFound        = errors.New("serializer not found")
	ErrCheckFieldSerializerNotSupported    = errors.New("serializer not supported")
	ErrCheckFieldInvalidFormat             = errors.New("invalid format")
	ErrTableNameNotCanonical               = errors.New("table name not canonical")
	ErrCheckFieldMutatorConflictPK         = errors.New("conflict mutators with primary_key")
	ErrCheckFieldMutatorConflictSerializer = errors.New("conflict mutators with serializer")
	ErrCheckFieldMutatorConflictObject     = errors.New("conflict mutators with object link")
	ErrCheckFieldSerializerConflictObject  = errors.New("conflict serializer with object link")
	ErrCheckServerEmpty                    = errors.New("serverConf is empty")
	ErrCheckFieldIndexEmpty                = errors.New("field for index is empty")
	ErrCheckObjectNotFound                 = errors.New("linked object not found")
	ErrCheckFieldTypeNotFound              = errors.New("procedure field type not found")
	ErrCheckFieldsEmpty                    = errors.New("empty required field declaration")
	ErrCheckFieldsManyDecl                 = errors.New("few declarations of fields not supported")
	ErrCheckFieldsProcNotImpl              = errors.New("proc fields not implemented")
	ErrCheckFieldsOrderDecl                = errors.New("incorrect order of fields")
)

// Описание ошибки декларации пакета
type ErrCheckPackageDecl struct {
	Pkg     string
	Backend activerecord.Backend
	Err     error
}

func (e *ErrCheckPackageDecl) Error() string {
	return ErrorBase(e)
}

// Описание ошибки декларации неймспейса
type ErrCheckPackageNamespaceDecl struct {
	Pkg  string
	Name string
	Err  error
}

func (e *ErrCheckPackageNamespaceDecl) Error() string {
	return ErrorBase(e)
}

// Описание ошибки декларации связанных сущностей
type ErrCheckPackageLinkedDecl struct {
	Pkg    string
	Object string
	Err    error
}

func (e *ErrCheckPackageLinkedDecl) Error() string {
	return ErrorBase(e)
}

// Описание ошибки декларации полей
type ErrCheckPackageFieldDecl struct {
	Pkg   string
	Field string
	Err   error
}

func (e *ErrCheckPackageFieldDecl) Error() string {
	return ErrorBase(e)
}

// Описание ошибки декларации мутаторов
type ErrCheckPackageFieldMutatorDecl struct {
	Pkg     string
	Field   string
	Mutator string
	Err     error
}

func (e *ErrCheckPackageFieldMutatorDecl) Error() string {
	return ErrorBase(e)
}

// Описание ошибки декларации индексов
type ErrCheckPackageIndexDecl struct {
	Pkg   string
	Index string
	Err   error
}

func (e *ErrCheckPackageIndexDecl) Error() string {
	return ErrorBase(e)
}
