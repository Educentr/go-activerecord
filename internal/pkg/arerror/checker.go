package arerror

import (
	"errors"
)

var (
	ErrCheckBackendEmpty                   = errors.New("backend empty")
	ErrCheckBackendUnknown                 = errors.New("backend unknown")
	ErrCheckEmptyNamespace                 = errors.New("empty namespace")
	ErrCheckPkgBackendToMatch              = errors.New("many backends for one class not supported yet")
	ErrCheckFieldSerializerNotFound        = errors.New("serializer not found")
	ErrCheckFieldSerializerNotSupported    = errors.New("serializer not supported")
	ErrCheckFieldInvalidFormat             = errors.New("invalid format")
	ErrCheckSelectAllNotSupported          = errors.New("select all not supported")
	ErrCheckFieldInvalidProcFormat         = errors.New("invalid proc format")
	ErrTableNameNotCanonical               = errors.New("table name not canonical. The general consensus is to use lowercase letters separated by underscores for readability and avoid reserved words to prevent confusion or errors")
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
	ErrCheckIndexConditionNotSupported        = errors.New("index condition not supported")
	ErrCheckIndexCountNotSupported            = errors.New("count by index not supported")
	ErrCheckIndexConditionHasNotValue         = errors.New("index condition without value")
	ErrCheckIndexConditionOperatorUnsupported = errors.New("operator not supported")
	ErrCheckIndexConditionValuesMismatch      = errors.New("operator value count mismatch")
	ErrCheckIndexConditionTypeIncompatible    = errors.New("field type incompatible with operator")
	ErrCheckIndexConditionNullCheckWithValues = errors.New("IS NULL/IS NOT NULL cannot have values")
	ErrCheckInternalError                     = errors.New("internal error")
)

// Описание ошибки декларации пакета
type ErrCheckPackageDecl struct {
	Pkg     string
	Backend string
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
