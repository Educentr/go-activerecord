package parser

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

const (
	indexFieldDeclCountProps = 2
)

func ParseIndexPartTag(field *ast.Field, ind *ds.IndexDeclaration, indexMap map[string]int, fields []ds.FieldDeclaration, indexes []ds.IndexDeclaration) error {
	tagParam, err := splitTag(field, CheckFlagEmpty, map[TagNameType]ParamValueRule{})
	if err != nil {
		return &arerror.ErrParseTypeIndexDecl{IndexType: "indexpart", Name: ind.Name, Err: err}
	}

	var exInd ds.IndexDeclaration

	var fieldnum int64 = 1

	for _, kv := range tagParam {
		switch TagNameType(kv[0]) {
		case SelectorTag:
			ind.Selector = kv[1]
		case "index":
			exIndNum, ex := indexMap[kv[1]]
			if !ex {
				return &arerror.ErrParseTypeIndexTagDecl{IndexType: "indexpart", Name: ind.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrIndexNotExist}
			}

			exInd = indexes[exIndNum]
			ind.Num = exInd.Num
		case "fieldnum":
			fNum, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return &arerror.ErrParseTypeIndexTagDecl{IndexType: "indexpart", Name: ind.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagValueInvalid}
			}

			fieldnum = fNum
		default:
			return &arerror.ErrParseTypeIndexTagDecl{IndexType: "indexpart", Name: ind.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrParseTagUnknown}
		}
	}

	if int64(len(exInd.Fields)) < fieldnum {
		return &arerror.ErrParseTypeIndexDecl{IndexType: "indexpart", Name: ind.Name, Err: arerror.ErrParseIndexFieldnumToBig}
	}

	if int64(len(exInd.Fields)) == fieldnum {
		return &arerror.ErrParseTypeIndexDecl{IndexType: "indexpart", Name: ind.Name, Err: arerror.ErrParseIndexFieldnumEqual}
	}

	for f := int64(0); f < fieldnum; f++ {
		ind.Fields = append(ind.Fields, exInd.Fields[f])
		ind.FieldsMap[fields[exInd.Fields[f]].Name] = exInd.FieldsMap[fields[exInd.Fields[f]].Name]
	}

	return nil
}

func ParseIndexPart(dst *ds.RecordPackage, fields []*ast.Field) error {
	if len(fields) == 0 {
		return nil
	}

	for _, field := range fields {
		if field.Names == nil || len(field.Names) != 1 {
			return &arerror.ErrParseTypeIndexDecl{IndexType: "indexpart", Err: arerror.ErrNameDeclaration}
		}

		ind := ds.IndexDeclaration{
			Name:      field.Names[0].Name,
			Fields:    []int{},
			FieldsMap: map[string]ds.IndexField{},
			Selector:  "SelectBy" + field.Names[0].Name,
			Partial:   true,
		}

		if err := checkBoolType(field.Type); err != nil {
			return &arerror.ErrParseTypeIndexDecl{IndexType: "indexpart", Name: ind.Name, Err: arerror.ErrTypeNotBool}
		}

		if err := ParseIndexPartTag(field, &ind, dst.IndexMap, dst.Fields, dst.Indexes); err != nil {
			return fmt.Errorf("error parse trigger tag: %w", err)
		}

		if len(ind.Fields) == 0 {
			return &arerror.ErrParseTypeIndexDecl{IndexType: "indexpart", Name: ind.Name, Err: arerror.ErrParseIndexFieldnumRequired}
		}

		if err := dst.AddIndex(ind); err != nil {
			return err
		}
	}

	return nil
}

func parseIndexConditionTag(condTag string, fieldsMap map[string]int) (map[int]ds.IndexCondition, *arerror.ErrParseTypeIndexTagDecl) {
	ret := map[int]ds.IndexCondition{}
	for _, cond := range strings.Split(condTag, ";") {
		start_cond := strings.Index(cond, "[")
		if start_cond == -1 {
			return nil, &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", TagValue: condTag, Err: arerror.ErrParseTagValueInvalid}
		}

		end_cond := strings.Index(cond[start_cond+1:], "]")
		if end_cond == -1 {
			return nil, &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", TagValue: condTag, Err: arerror.ErrParseTagValueInvalid}
		}

		if len(cond) == start_cond+end_cond+2 {
			return nil, &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", TagValue: condTag, Err: arerror.ErrParseTagValueInvalid}
		}

		fieldName := cond[:start_cond]
		if fldNum, ex := fieldsMap[fieldName]; !ex {
			return nil, &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", TagValue: condTag, Err: arerror.ErrFieldNotExist}
		} else if _, ex := ret[fldNum]; ex {
			return nil, &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", TagValue: condTag, Err: arerror.ErrDuplicate}
		} else {
			ret[fldNum] = ds.IndexCondition{ConditionType: cond[start_cond+1 : start_cond+end_cond+1], Value: strings.Split(cond[start_cond+end_cond+2:], ",")}
		}
	}

	return ret, nil
}

// ToDo объединить с парсингом частично индекса, есть повторяющийся код!
func ParseIndexTag(field *ast.Field, ind *ds.IndexDeclaration, fieldsMap map[string]int) error {
	tagParam, err := splitTag(field, CheckFlagEmpty, map[TagNameType]ParamValueRule{PrimaryKeyTag: ParamNotNeedValue, UniqueTag: ParamNotNeedValue})
	if err != nil {
		return &arerror.ErrParseTypeIndexDecl{IndexType: "index", Name: ind.Name, Err: err}
	}

	for _, kv := range tagParam {
		switch TagNameType(kv[0]) {
		case PrimaryKeyTag:
			ind.Primary = true
		case UniqueTag:
			ind.Unique = true
		case SelectorTag:
			ind.Selector = kv[1]
		case ConditionalTag:
			var errCond *arerror.ErrParseTypeIndexTagDecl

			ind.Conditions, errCond = parseIndexConditionTag(kv[1], fieldsMap)
			if errCond != nil {
				errCond.Name = ind.Name
				errCond.TagName = kv[0]

				return errCond
			}
		case FieldsTag:
			for _, fieldDecl := range strings.Split(kv[1], ",") {
				fieldDeclPars := strings.SplitN(fieldDecl, "=", indexFieldDeclCountProps)
				fieldName := fieldDeclPars[0]

				fieldOrder := ds.IndexOrderAsc
				if len(fieldDeclPars) > 1 && fieldDeclPars[1] == "desc" {
					fieldOrder = ds.IndexOrderDesc
				}

				if fldNum, ex := fieldsMap[fieldName]; !ex {
					return &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", Name: ind.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrFieldNotExist}
				} else if _, ex := ind.FieldsMap[fieldName]; ex {
					return &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", Name: ind.Name, TagName: kv[0], TagValue: kv[1], Err: arerror.ErrDuplicate}
				} else {
					ind.FieldsMap[fieldName] = ds.IndexField{IndField: fldNum, Order: fieldOrder}
					ind.Fields = append(ind.Fields, fldNum)
				}
			}
		// case "selector_count":
		// 	ind.SelectorCount = "CountBy" + ind.Name

		// 	if kv[1] != "" {
		// 		ind.SelectorCount = kv[1]
		// 	}
		default:
			val := "NO VALUE"
			if len(kv) > 1 {
				val = kv[1]
			}

			return &arerror.ErrParseTypeIndexTagDecl{IndexType: "index", Name: ind.Name, TagName: kv[0], TagValue: val, Err: arerror.ErrParseTagUnknown}
		}
	}

	return nil
}

func ParseIndexes(dst *ds.RecordPackage, fields []*ast.Field) error {
	if len(fields) == 0 {
		return nil
	}

	for _, field := range fields {
		if field.Names == nil || len(field.Names) != 1 {
			return &arerror.ErrParseTypeIndexDecl{IndexType: "index", Err: arerror.ErrNameDeclaration}
		}

		ind := ds.IndexDeclaration{
			Name:       field.Names[0].Name,
			Fields:     []int{},
			FieldsMap:  map[string]ds.IndexField{},
			Selector:   "SelectBy" + field.Names[0].Name,
			Conditions: map[int]ds.IndexCondition{},
		}
		if err := checkBoolType(field.Type); err != nil {
			return &arerror.ErrParseTypeIndexDecl{IndexType: "index", Name: ind.Name, Err: arerror.ErrTypeNotBool}
		}

		if err := ParseIndexTag(field, &ind, dst.FieldsMap); err != nil {
			return fmt.Errorf("error parse indexTag: %w", err)
		}

		if err := dst.AddIndex(ind); err != nil {
			return err
		}
	}

	return nil
}
