package checker

import (
	"reflect"
	"testing"

	"github.com/Educentr/go-activerecord/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/pkg/activerecord"
	"github.com/Educentr/go-activerecord/pkg/octopus"
)

func TestCheck(t *testing.T) {
	rpFoo := ds.NewRecordPackage()
	rpFoo.Backends = []activerecord.Backend{octopus.Backend}
	rpFoo.Namespace = ds.NamespaceDeclaration{ObjectName: "0", PackageName: "foo", PublicName: "Foo"}
	rpFoo.ServerConfKey = "testCheckerConfKey"

	err := rpFoo.AddField(ds.FieldDeclaration{
		Name:       "ID",
		Format:     octopus.Int,
		PrimaryKey: true,
		Mutators:   []string{},
		Size:       0,
		Serializer: []string{},
		ObjectLink: "",
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	err = rpFoo.AddField(ds.FieldDeclaration{
		Name:       "BarID",
		Format:     octopus.Int,
		PrimaryKey: false,
		Mutators:   []string{},
		Size:       0,
		Serializer: []string{},
		ObjectLink: "Bar",
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	err = rpFoo.AddFieldObject(ds.FieldObject{
		Name:       "Foo",
		Key:        "ID",
		ObjectName: "bar",
		Field:      "BarID",
		Unique:     true,
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	rpInvalidFormat := ds.NewRecordPackage()
	rpInvalidFormat.Backends = []activerecord.Backend{octopus.Backend}
	rpInvalidFormat.Namespace = ds.NamespaceDeclaration{ObjectName: "0", PackageName: "invform", PublicName: "InvalidFormat"}
	rpInvalidFormat.ServerConfKey = "testCheckerConfKey"

	err = rpInvalidFormat.AddField(ds.FieldDeclaration{
		Name:       "ID",
		Format:     "byte",
		PrimaryKey: true,
		Mutators:   []string{},
		Size:       0,
		Serializer: []string{},
		ObjectLink: "",
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	onInvalidFormat := ds.NewRecordPackage()
	onInvalidFormat.Backends = []activerecord.Backend{octopus.Backend}
	onInvalidFormat.Namespace = ds.NamespaceDeclaration{ObjectName: "invalid", PackageName: "invform", PublicName: "InvalidFormat"}
	rpInvalidFormat.ServerConfKey = "testCheckerConfKey"

	err = onInvalidFormat.AddField(ds.FieldDeclaration{
		Name:       "ID",
		Format:     "byte",
		PrimaryKey: true,
		Mutators:   []string{},
		Size:       0,
		Serializer: []string{},
		ObjectLink: "",
	})
	if err != nil {
		t.Errorf("can't prepare test data: %s", err)
		return
	}

	type args struct {
		files         map[string]*ds.RecordPackage
		linkedObjects map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "octopus empty",
			args: args{
				files:         map[string]*ds.RecordPackage{},
				linkedObjects: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "linked objs",
			args: args{
				files:         map[string]*ds.RecordPackage{"foo": rpFoo},
				linkedObjects: map[string]string{"bar": "bar"},
			},
			wantErr: false,
		},
		{
			name: "wrong octopus format",
			args: args{
				files:         map[string]*ds.RecordPackage{"invalid": rpInvalidFormat},
				linkedObjects: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "wrong octopus namespace objectname format",
			args: args{
				files:         map[string]*ds.RecordPackage{"invalid": onInvalidFormat},
				linkedObjects: map[string]string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Check(tt.args.files, tt.args.linkedObjects); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInit(t *testing.T) {
	type args struct {
		files map[string]*ds.RecordPackage
	}
	tests := []struct {
		name string
		args args
		want *Checker
	}{
		{
			name: "simple init",
			args: args{
				files: map[string]*ds.RecordPackage{},
			},
			want: &Checker{
				files: map[string]*ds.RecordPackage{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Init(tt.args.files); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}
