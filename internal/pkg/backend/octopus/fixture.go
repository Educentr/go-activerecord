package octopus

// ToDo merge with postgres backend
import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/textutil"
)

type FixturePkgData struct {
	FixturePkg       string
	ARPkg            string
	ARPkgTitle       string
	FieldList        []ds.FieldDeclaration
	FieldMap         map[string]int
	FieldObject      map[string]ds.FieldObject
	ProcInFieldList  []ds.ProcFieldDeclaration
	ProcOutFieldList []ds.ProcFieldDeclaration
	Container        ds.NamespaceDeclaration
	Indexes          []ds.IndexDeclaration
	Serializers      map[string]ds.SerializerDeclaration
	Mutators         map[string]ds.MutatorDeclaration
	Imports          []ds.ImportDeclaration
	AppInfo          string
}

//go:embed tmpl/fixturestore.tmpl
var FixtureTmpl string

var FixtureTemplateFuncs = template.FuncMap{"snakeCase": textutil.ToSnakeCase, "split": strings.Split}
