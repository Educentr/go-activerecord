package checker

import (
	"testing"

	"github.com/mailru/activerecord/internal/pkg/ds"
	"github.com/mailru/activerecord/pkg/activerecord"
	"github.com/mailru/activerecord/pkg/octopus"
	"github.com/mailru/activerecord/pkg/postgres"
)

func Test_checkBackend(t *testing.T) {
	rcOctopus := ds.NewRecordPackage()
	rcOctopus.Backends = []activerecord.Backend{octopus.Backend}
	rcMany := ds.NewRecordPackage()
	rcMany.Backends = []activerecord.Backend{octopus.Backend, postgres.Backend}

	type args struct {
		cl *ds.RecordPackage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "emptyBack", args: args{cl: ds.NewRecordPackage()}, wantErr: true},
		{name: "oneBack", args: args{cl: rcOctopus}, wantErr: false},
		{name: "manyBack", args: args{cl: rcMany}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkBackend(tt.args.cl); (err != nil) != tt.wantErr {
				t.Errorf("CheckBackend() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkLinkedObject(t *testing.T) {
	rp := ds.NewRecordPackage()
	rpLinked := ds.NewRecordPackage()

	err := rpLinked.AddFieldObject(ds.FieldObject{
		Name:       "Foo",
		Key:        "ID",
		ObjectName: "bar",
		Field:      "barID",
		Unique:     false,
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	type args struct {
		cl            *ds.RecordPackage
		linkedObjects map[string]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "without linked obj", args: args{cl: rp, linkedObjects: map[string]string{}}, wantErr: false},
		{name: "no linked obj", args: args{cl: rpLinked, linkedObjects: map[string]string{}}, wantErr: true},
		{name: "normal linked obj", args: args{cl: rpLinked, linkedObjects: map[string]string{"bar": "bar"}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkLinkedObject(tt.args.cl, tt.args.linkedObjects); (err != nil) != tt.wantErr {
				t.Errorf("checkLinkedObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkNamespace(t *testing.T) {
	type args struct {
		ns ds.NamespaceDeclaration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal namespace",
			args: args{
				ns: ds.NamespaceDeclaration{
					ObjectName:  "0",
					PublicName:  "Foo",
					PackageName: "foo",
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			args: args{
				ns: ds.NamespaceDeclaration{
					ObjectName:  "0",
					PublicName:  "",
					PackageName: "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "empty package",
			args: args{
				ns: ds.NamespaceDeclaration{
					ObjectName:  "0",
					PublicName:  "Foo",
					PackageName: "",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkNamespace(tt.args.ns); (err != nil) != tt.wantErr {
				t.Errorf("checkNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkFields(t *testing.T) {
	type args struct {
		cl ds.RecordPackage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty fields",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{},
				},
			},
			wantErr: true,
		},
		{
			name: "empty format",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name: "Foo",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:   "Foo",
							Format: "[]int",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no primary",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:   "Foo",
							Format: "int",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "normal field",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "fields conflict with links",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
						},
					},
					FieldsObjectMap: map[string]ds.FieldObject{
						"Foo": {},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "mutators and primary",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
							Mutators: []string{
								"fmut",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "serializer not declared",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
						},
						{
							Name:   "Foo",
							Format: "int",
							Mutators: []string{
								"fmut",
							},
							Serializer: []string{
								"fser",
							},
						},
					},
					SerializerMap: map[string]ds.SerializerDeclaration{},
				},
			},
			wantErr: true,
		},
		{
			name: "mutators and links",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
						},
						{
							Name:   "Foo",
							Format: "int",
							Mutators: []string{
								"fmut",
							},
							ObjectLink: "Bar",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "serializer and links",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
						},
						{
							Name:   "Foo",
							Format: "int",
							Serializer: []string{
								"fser",
							},
							ObjectLink: "Bar",
						},
					},
					SerializerMap: map[string]ds.SerializerDeclaration{
						"fser": {},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "custom mutator without serializer",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Pk",
							Format:     "int",
							PrimaryKey: true,
						},
						{
							Name:   "Foo",
							Format: "string",
							Mutators: []string{
								"cmut",
							},
						},
					},
					MutatorMap: map[string]ds.MutatorDeclaration{
						"cmut": {
							Name:          "cmut",
							Type:          "pkg.Bar",
							PartialFields: make([]ds.PartialFieldDeclaration, 1),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "few custom mutator on field",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Pk",
							Format:     "int",
							PrimaryKey: true,
						},
						{
							Name:   "Foo",
							Format: "string",
							Mutators: []string{
								"dec", "cmut", "cmut2",
							},
						},
					},
					MutatorMap: map[string]ds.MutatorDeclaration{
						"cmut": {
							Name: "cmut",
							Type: "string",
						},
						"cmut2": {
							Name: "cmut2",
							Type: "string",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFields(&tt.args.cl); (err != nil) != tt.wantErr {
				t.Errorf("checkFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkProcFields(t *testing.T) {
	type args struct {
		cl ds.RecordPackage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty fields",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{},
				},
			},
			wantErr: true,
		},
		{
			name: "2 fields declaration",
			args: args{
				cl: ds.RecordPackage{
					Fields: []ds.FieldDeclaration{
						{
							Name:       "Foo",
							Format:     "int",
							PrimaryKey: true,
						},
					},
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
							Type:   ds.INOUT,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty format",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name: "Foo",
							Type: ds.OUT,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid input format",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
							Type:   ds.OUT,
						},
					},
					ProcInFields: []ds.ProcFieldDeclaration{
						{
							Name:   "Foo",
							Format: "[]int",
							Type:   ds.IN,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid output format",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "[]int",
							Type:   ds.OUT,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "type not found",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "incorrect fields order",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
						},
						2: {
							Name:   "Bar",
							Format: "int",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "normal field",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
							Type:   ds.OUT,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal input field",
			args: args{
				cl: ds.RecordPackage{
					ProcInFields: []ds.ProcFieldDeclaration{
						{
							Name:   "Foo",
							Format: "[]string",
							Type:   ds.IN,
						},
					},
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
							Type:   ds.OUT,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "serializer not declared",
			args: args{
				cl: ds.RecordPackage{
					ProcOutFields: ds.ProcFieldDeclarations{
						0: {
							Name:   "Foo",
							Format: "int",
							Type:   ds.OUT,
						},
						1: {
							Name:   "Foo",
							Format: "int",
							Type:   ds.OUT,
							Serializer: []string{
								"fser",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFields(&tt.args.cl); (err != nil) != tt.wantErr {
				t.Errorf("checkFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
