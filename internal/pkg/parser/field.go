package parser

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/Educentr/go-activerecord/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/pkg/activerecord"
)

// Функция парсинга тегов полей модели
func ParseFieldsTag(field *ast.Field, newfield *ds.FieldDeclaration, newindex *ds.IndexDeclaration) error {
	tagParam, err := splitTag(field, NoCheckFlag, map[TagNameType]ParamValueRule{PrimaryKeyTag: ParamNotNeedValue, UniqueTag: ParamNotNeedValue})
	if err != nil {
		return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: string(newfield.Format), Err: err}
	}

	if len(tagParam) > 0 {
		for _, kv := range tagParam {
			switch TagNameType(kv[0]) {
			case SelectorTag:
				newindex.Name = newfield.Name
				newindex.Selector = kv[1]
			case PrimaryKeyTag:
				newindex.Name = newfield.Name
				newindex.Primary = true
			case UniqueTag:
				newindex.Name = newfield.Name
				newindex.Unique = true
			case MutatorsTag:
				newfield.Mutators = strings.Split(kv[1], ",")
			case SizeTag:
				if kv[1] != "" {
					size, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						return &arerror.ErrParseTypeFieldTagDecl{Name: newfield.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagValueInvalid}
					}

					newfield.Size = size
				}
			case SerializerTag:
				newfield.Serializer = strings.Split(kv[1], ",")
			case InitByDBTag:
				newfield.InitByDB = true
			default:
				return &arerror.ErrParseTypeFieldTagDecl{Name: newfield.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagUnknown}
			}
		}
	}

	return nil
}

// Функция парсинга полей модели
func ParseFields(dst *ds.RecordPackage, fields []*ast.Field) error {
	for _, field := range fields {
		if field.Names == nil || len(field.Names) != 1 {
			return &arerror.ErrParseTypeFieldDecl{Err: arerror.ErrNameDeclaration}
		}

		newfield := ds.FieldDeclaration{
			Name:       field.Names[0].Name,
			Mutators:   []string{},
			Serializer: []string{},
		}

		newindex := ds.IndexDeclaration{
			FieldsMap: map[string]ds.IndexField{},
		}

		switch t := field.Type.(type) {
		case *ast.Ident:
			newfield.Format = activerecord.Format(t.String())
		case *ast.ArrayType:
			//Todo точно ли массив надо, а не срез?
			if t.Elt.(*ast.Ident).Name != "byte" {
				return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: t.Elt.(*ast.Ident).Name, Err: arerror.ErrParseFieldArrayOfNotByte}
			}

			if t.Len == nil {
				return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: t.Elt.(*ast.Ident).Name, Err: arerror.ErrParseFieldArrayNotSlice}
			}

			return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: t.Elt.(*ast.Ident).Name, Err: arerror.ErrParseFieldBinary}
		default:
			return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: fmt.Sprintf("%T", t), Err: arerror.ErrUnknown}
		}

		if err := ParseFieldsTag(field, &newfield, &newindex); err != nil {
			return fmt.Errorf("error ParseFieldsTag: %w", err)
		}

		if err := dst.AddField(newfield); err != nil {
			return err
		}

		if newindex.Name != "" {
			newindex.Fields = append(newindex.Fields, len(dst.Fields)-1)
			newindex.FieldsMap[newfield.Name] = ds.IndexField{IndField: len(dst.Fields) - 1, Order: ds.IndexOrderAsc}

			errIndex := dst.AddIndex(newindex)
			if errIndex != nil {
				return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: string(newfield.Format), Err: errIndex}
			}
		}
	}

	return nil
}

// ParseProcFieldsTag парсинг тегов полей декларации процедуры
func ParseProcFieldsTag(index int, field *ast.Field, newfield *ds.ProcFieldDeclaration) error {
	tagParam, err := splitTag(field, NoCheckFlag, map[TagNameType]ParamValueRule{PrimaryKeyTag: ParamNotNeedValue, UniqueTag: ParamNotNeedValue})
	if err != nil {
		return &arerror.ErrParseTypeFieldDecl{Name: newfield.Name, FieldType: string(newfield.Format), Err: err}
	}

	if len(tagParam) > 0 {
		for _, kv := range tagParam {
			switch TagNameType(kv[0]) {
			case ProcInputParamTag:
				//результат бинарной операции 0|IN => IN; 1|IN => IN; 2|IN => INOUT (3);
				newfield.Type = newfield.Type | ds.IN
			case ProcOutputParamTag:
				//результат бинарной операции 0|OUT => OUT; 1|OUT => INOUT (3); 2|OUT => OUT;
				newfield.Type = newfield.Type | ds.OUT
				orderIdx := index

				if len(kv) == 2 {
					orderIdx, err = strconv.Atoi(kv[1])
					if err != nil {
						return &arerror.ErrParseTypeFieldTagDecl{Name: newfield.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagValueInvalid}
					}
				}

				newfield.OrderIndex = orderIdx
			case SizeTag:
				if kv[1] != "" {
					size, err := strconv.ParseInt(kv[1], 10, 64)
					if err != nil {
						return &arerror.ErrParseTypeFieldTagDecl{Name: newfield.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagValueInvalid}
					}

					newfield.Size = size
				}
			case SerializerTag:
				newfield.Serializer = strings.Split(kv[1], ",")
			default:
				return &arerror.ErrParseTypeFieldTagDecl{Name: newfield.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagUnknown}
			}
		}
	}

	return nil
}

// ParseProcFields парсинг полей процедуры
func ParseProcFields(dst *ds.RecordPackage, fields []*ast.Field) error {
	for _, field := range fields {
		if field.Names == nil || len(field.Names) != 1 {
			return &arerror.ErrParseTypeFieldDecl{Err: arerror.ErrNameDeclaration}
		}

		newField := ds.ProcFieldDeclaration{
			Name:       field.Names[0].Name,
			Serializer: []string{},
		}

		if err := ParseProcFieldsTag(len(dst.ProcOutFields), field, &newField); err != nil {
			return fmt.Errorf("error ParseFieldsTag: %w", err)
		}

		switch t := field.Type.(type) {
		case *ast.Ident:
			newField.Format = activerecord.Format(t.String())
		case *ast.ArrayType:
			if t.Elt.(*ast.Ident).Name != "byte" && t.Elt.(*ast.Ident).Name != "string" {
				return &arerror.ErrParseTypeFieldDecl{Name: newField.Name, FieldType: t.Elt.(*ast.Ident).Name, Err: arerror.ErrParseProcFieldArraySlice}
			}

			// если входной параметр slice
			if newField.Type == ds.IN && t.Len == nil {
				newField.Format = activerecord.Format(fmt.Sprintf("[]%s", t.Elt.(*ast.Ident).Name))
				break
			}

			if t.Len == nil {
				return &arerror.ErrParseTypeFieldDecl{Name: newField.Name, FieldType: t.Elt.(*ast.Ident).Name, Err: arerror.ErrParseFieldArrayNotSlice}
			}

			return &arerror.ErrParseTypeFieldDecl{Name: newField.Name, FieldType: t.Elt.(*ast.Ident).Name, Err: arerror.ErrParseFieldBinary}
		default:
			return &arerror.ErrParseTypeFieldDecl{Name: newField.Name, FieldType: fmt.Sprintf("%T", t), Err: arerror.ErrUnknown}
		}

		if err := dst.AddProcField(newField); err != nil {
			return err
		}
	}

	return nil
}
