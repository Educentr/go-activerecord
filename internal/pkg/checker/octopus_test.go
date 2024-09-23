package checker

import (
	"testing"

	"github.com/Educentr/go-activerecord/internal/pkg/arerror"
	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/pkg/activerecord"
	"github.com/Educentr/go-activerecord/pkg/octopus"
	"github.com/stretchr/testify/assert"
)

func TestCheckFields(t *testing.T) {
	checker := CreateOctopusChecker().(*octopusChecker)

	tests := []struct {
		name    string
		record  *ds.RecordPackage
		wantErr error
	}{
		{
			name: "Fields and ProcOutFields both present",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				Fields:    []ds.FieldDeclaration{{Format: activerecord.Format("int")}},
				ProcOutFields: ds.ProcFieldDeclarations{
					1: ds.ProcFieldDeclaration{Format: activerecord.Format("int")},
				},
			},
			wantErr: &arerror.ErrCheckPackageDecl{Pkg: "testpkg", Err: arerror.ErrCheckFieldsManyDecl},
		},
		{
			name: "Invalid field format",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				Fields:    []ds.FieldDeclaration{{Format: activerecord.Format("float")}},
			},
			wantErr: &arerror.ErrCheckPackageFieldDecl{Pkg: "testpkg", Field: "", Err: arerror.ErrCheckFieldInvalidFormat},
		},
		{
			name: "Empty field format",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				Fields:    []ds.FieldDeclaration{},
			},
			wantErr: &arerror.ErrCheckPackageDecl{Pkg: "testpkg", Backend: "", Err: arerror.ErrCheckFieldsEmpty},
		},
		{
			name: "No fields and no ProcOutFields",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
			},
			wantErr: &arerror.ErrCheckPackageDecl{Pkg: "testpkg", Err: arerror.ErrCheckFieldsEmpty},
		},
		{
			name: "Valid fields",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				Fields:    []ds.FieldDeclaration{{Format: octopus.String}},
			},
			wantErr: nil,
		},
		{
			name: "invalid input format",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				ProcOutFields: ds.ProcFieldDeclarations{
					0: ds.ProcFieldDeclaration{Format: activerecord.Format("int"), Name: "Foo", Type: ds.OUT},
				},
				ProcInFields: []ds.ProcFieldDeclaration{
					{
						Name:   "Foo",
						Format: "[]int",
						Type:   ds.IN,
					},
				},
			},
			wantErr: &arerror.ErrCheckPackageFieldDecl{Pkg: "testpkg", Field: "Foo", Err: arerror.ErrCheckFieldInvalidFormat},
		},
		{
			name: "invalid output format",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				ProcOutFields: ds.ProcFieldDeclarations{
					0: ds.ProcFieldDeclaration{Format: activerecord.Format("[]int"), Name: "Foo", Type: ds.OUT},
				},
			},
			wantErr: &arerror.ErrCheckPackageFieldDecl{Pkg: "testpkg", Field: "Foo", Err: arerror.ErrCheckFieldInvalidFormat},
		},
		{
			name: "empty type",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				ProcOutFields: ds.ProcFieldDeclarations{
					0: ds.ProcFieldDeclaration{Format: activerecord.Format("int"), Name: "Foo"},
				},
			},
			wantErr: &arerror.ErrCheckPackageFieldDecl{Pkg: "testpkg", Field: "Foo", Err: arerror.ErrCheckFieldTypeNotFound},
		},
		{
			name: "incorrect fields order",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				ProcOutFields: ds.ProcFieldDeclarations{
					0: ds.ProcFieldDeclaration{Format: activerecord.Format("int"), Name: "Foo"},
					2: ds.ProcFieldDeclaration{Format: activerecord.Format("int"), Name: "Bar"},
				},
			},
			wantErr: &arerror.ErrCheckPackageDecl{Pkg: "testpkg", Err: arerror.ErrCheckFieldsOrderDecl},
		},
		{
			name: "serializer not declared",
			record: &ds.RecordPackage{
				Namespace: ds.NamespaceDeclaration{PackageName: "testpkg"},
				ProcOutFields: ds.ProcFieldDeclarations{
					0: ds.ProcFieldDeclaration{Format: activerecord.Format("int"), Name: "Foo", Type: ds.OUT},
					1: ds.ProcFieldDeclaration{Format: activerecord.Format("int"), Name: "Foo", Type: ds.OUT, Serializer: []string{"fser"}},
				},
			},
			wantErr: &arerror.ErrCheckPackageFieldDecl{Pkg: "testpkg", Field: "Foo", Err: arerror.ErrCheckFieldSerializerNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checker.checkFields(tt.record)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
