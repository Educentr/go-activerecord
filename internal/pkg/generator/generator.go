package generator

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/tools/imports"

	"github.com/mailru/activerecord/internal/pkg/arerror"
	"github.com/mailru/activerecord/internal/pkg/ds"
	"github.com/mailru/activerecord/pkg/activerecord"
	"github.com/mailru/activerecord/pkg/octopus"
	"github.com/mailru/activerecord/pkg/postgres"
)

const disclaimer string = `// Code generated by argen. DO NOT EDIT.
// This code was generated from a template.
//
// Manual changes to this file may cause unexpected behavior in your application.
// Manual changes to this file will be overwritten if the code is regenerated.
//
// Generate info: {{ .AppInfo }}
`

type PkgData struct {
	ARPkg            string
	ARPkgTitle       string
	FieldList        []ds.FieldDeclaration
	FieldMap         map[string]int
	FieldObject      map[string]ds.FieldObject
	LinkedObject     map[string]ds.RecordPackage
	ProcInFieldList  []ds.ProcFieldDeclaration
	ProcOutFieldList []ds.ProcFieldDeclaration
	ServerConfKey    string
	Container        ds.NamespaceDeclaration
	Indexes          []ds.IndexDeclaration
	Serializers      map[string]ds.SerializerDeclaration
	Mutators         map[string]ds.MutatorDeclaration
	Imports          []ds.ImportDeclaration
	Triggers         map[string]ds.TriggerDeclaration
	Flags            map[string]ds.FlagDeclaration
	AppInfo          string
}

func NewPkgData(appInfo string, cl ds.RecordPackage) PkgData {
	return PkgData{
		ARPkg:            cl.Namespace.PackageName,
		ARPkgTitle:       cl.Namespace.PublicName,
		Indexes:          cl.Indexes,
		FieldList:        cl.Fields,
		FieldMap:         cl.FieldsMap,
		ProcInFieldList:  cl.ProcInFields,
		ProcOutFieldList: cl.ProcOutFields.List(),
		FieldObject:      cl.FieldsObjectMap,
		ServerConfKey:    cl.ServerConfKey,
		Container:        cl.Namespace,
		Serializers:      cl.SerializerMap,
		Mutators:         cl.MutatorMap,
		Imports:          cl.Imports,
		Triggers:         cl.TriggerMap,
		Flags:            cl.FlagMap,
		AppInfo:          appInfo,
	}
}

const TemplateName = `ARPkgTemplate`

type GenerateFile struct {
	Data    []byte
	Name    string
	Dir     string
	Backend activerecord.Backend
}

type MetaData struct {
	Namespaces []*ds.RecordPackage
	AppInfo    string
}

//nolint:revive
//go:embed tmpl/meta.tmpl
var MetaTmpl string

func GenerateMeta(params MetaData) ([]GenerateFile, *arerror.ErrGeneratorFile) {
	genData, aeErr := GenerateByTmpl(params, "meta", OctopusTemplateFuncs, "meta.tmpl", MetaTmpl)
	if aeErr != nil {
		return nil, &arerror.ErrGeneratorFile{Name: "repository.go", Backend: "meta", Filename: "repository.go", Err: aeErr} //ToDo error assertion
	}

	genRes := GenerateFile{
		Dir:     "",
		Name:    "repository.go",
		Backend: "meta",
	}

	var err error

	genRes.Data, err = imports.Process("", genData, nil)
	if err != nil {
		return nil, &arerror.ErrGeneratorFile{Name: "repository.go", Backend: "meta", Filename: genRes.Name, Err: ErrorLine(err, string(genData))}
	}

	return []GenerateFile{genRes}, nil
}

func GenerateByTmpl(params any, backendName string, funcs template.FuncMap, templateName, tmpl string) ([]byte, *arerror.ErrGeneratorPhases) {
	templatePackage, err := template.New(TemplateName).Funcs(funcs).Funcs(BaseTemplateFuncs).Funcs(OctopusTemplateFuncs).Parse(tmpl)
	if err != nil {
		tmplLines, errgetline := getTmplErrorLine(strings.SplitAfter(tmpl, "\n"), err.Error())
		if errgetline != nil {
			tmplLines = errgetline.Error()
		}

		return nil, &arerror.ErrGeneratorPhases{Backend: backendName, Phase: "parse", TmplLines: tmplLines, Err: err}
	}

	buf := &bytes.Buffer{}
	writer := bufio.NewWriter(buf)

	err = templatePackage.Execute(writer, params)
	if err != nil {
		tmplLines, errgetline := getTmplErrorLine(strings.SplitAfter(tmpl, "\n"), err.Error())
		if errgetline != nil {
			tmplLines = errgetline.Error()
		}

		return nil, &arerror.ErrGeneratorPhases{Backend: backendName, Phase: "execute", TmplLines: tmplLines, Err: err}
	}

	err = writer.Flush()
	if err != nil {
		return nil, &arerror.ErrGeneratorPhases{Backend: backendName, Name: templateName, Phase: "generate", Err: err}
	}

	pkgContentBytes := buf.Bytes()
	pkgContent, err := imports.Process("", pkgContentBytes, nil)
	retStr := disclaimer + string(pkgContent)

	if err != nil {
		return nil, &arerror.ErrGeneratorPhases{Backend: backendName, Name: templateName, Phase: "import", Err: ErrorLine(err, string(pkgContentBytes))}
	}

	rxPkg := regexp.MustCompile(`(?m)^package .*$`)
	rxSpace := regexp.MustCompile(`\s`)

	afterRx := rxSpace.ReplaceAll(rxPkg.ReplaceAll(pkgContent, []byte{}), []byte{})

	if len(afterRx) == 0 {
		return []byte{}, nil
	}

	return []byte(retStr), nil
}

func Generate(appInfo string, cl ds.RecordPackage, linkObject map[string]ds.RecordPackage) (ret []GenerateFile, err error) {
	for _, backend := range cl.Backends {
		var generated map[string][]byte

		params := NewPkgData(appInfo, cl)
		params.LinkedObject = linkObject

		//log.Printf("Generate package (%v)", cl)

		var err *arerror.ErrGeneratorPhases

		switch backend {
		case octopus.BackendTarantool:
			fallthrough
		case octopus.Backend:
			generated, err = GenerateFromDir(params, "octopus", OctopusTemplatesPath, tmplOctopusPath, OctopusTemplateFuncs)
		case postgres.Backend:
			generated, err = GenerateFromDir(params, "postgres", postgresTemplatesPath, tmplPostgresPath, PostgresTemplateFuncs)
		case "tarantool16":
			fallthrough
		case "tarantool2":
			return nil, &arerror.ErrGeneratorFile{Name: cl.Namespace.PublicName, Backend: backend, Err: arerror.ErrGeneratorBackendNotImplemented}
		default:
			return nil, &arerror.ErrGeneratorFile{Name: cl.Namespace.PublicName, Backend: backend, Err: arerror.ErrGeneratorBackendUnknown}
		}

		if err != nil {
			err.Name = cl.Namespace.PublicName
			return nil, err
		}

		for name, genData := range generated {
			genRes := GenerateFile{
				Dir:     cl.Namespace.PackageName,
				Name:    name + ".go",
				Backend: backend,
				Data:    genData,
			}

			ret = append(ret, genRes)
		}
	}

	return ret, nil
}

func GenerateFromDir(params PkgData, backendName string, embedFS embed.FS, tmplPath string, funcs template.FuncMap) (map[string][]byte, *arerror.ErrGeneratorPhases) {
	templates, err := embedFS.ReadDir(tmplPath)
	if err != nil {
		return nil, &arerror.ErrGeneratorPhases{Backend: backendName, Phase: "generate", Err: err}
	}

	ret := make(map[string][]byte, len(templates))

	for _, template := range templates {
		if !strings.HasSuffix(template.Name(), ".tmpl") {
			if !template.IsDir() {
				log.Printf("%s has no suffix tmpl. skip", template.Name())
			}

			continue
		}

		tmpl, err := embedFS.ReadFile(path.Join(tmplPath, template.Name()))
		if err != nil {
			return nil, &arerror.ErrGeneratorPhases{Backend: backendName, Name: template.Name(), Phase: "generate", Err: err}
		}

		fileName := template.Name()[:len(template.Name())-5]

		data, aeErr := GenerateByTmpl(params, backendName, funcs, template.Name(), string(tmpl))
		if aeErr != nil {
			return nil, aeErr
		}

		if len(data) > 0 {
			ret[fileName] = data
		}
	}

	return ret, nil
}

var errImportsRx = regexp.MustCompile(`^(\d+):(\d+):`)

func ErrorLine(errIn error, genData string) error {
	findErr := errImportsRx.FindStringSubmatch(errIn.Error())
	if len(findErr) == 3 {
		lineNum, err := strconv.Atoi(findErr[1])
		if err != nil {
			return errors.Wrap(errIn, "can't parse error line num")
		}

		lines := strings.Split(genData, "\n")

		if len(lines) < lineNum {
			return errors.Wrap(errIn, fmt.Sprintf("line num %d not found (total %d)", lineNum, len(lines)))
		}

		line := lines[lineNum-1]

		byteNum, err := strconv.Atoi(findErr[2])
		if err != nil {
			return errors.Wrap(errIn, "can't parse error byte num in line: "+line)
		}

		if len(line)+1 < byteNum {
			return errors.Wrap(errIn, "byte num not found in line: "+line)
		}

		strs := "\n"
		for i := -10; i < -1; i++ {
			strs += strings.Trim(lines[lineNum+i], "\t") + "\n"
		}

		return errors.Wrap(errIn, strs+strings.Trim(line, "\t")+"\n"+strings.Repeat(" ", byteNum-1)+"^^^^^"+"\n"+strings.Trim(lines[lineNum], "\t"))
	}

	return errors.Wrap(errIn, "can't parse error message")
}

func GenerateFixture(appInfo string, cl ds.RecordPackage, pkg string, pkgFixture string) ([]GenerateFile, error) {
	var generated map[string]bytes.Buffer

	ret := make([]GenerateFile, 0, 1)

	params := FixturePkgData{
		FixturePkg:       pkgFixture,
		ARPkg:            pkg,
		ARPkgTitle:       cl.Namespace.PublicName,
		FieldList:        cl.Fields,
		FieldMap:         cl.FieldsMap,
		FieldObject:      cl.FieldsObjectMap,
		ProcInFieldList:  cl.ProcInFields,
		ProcOutFieldList: cl.ProcOutFields.List(),
		Container:        cl.Namespace,
		Indexes:          cl.Indexes,
		Serializers:      cl.SerializerMap,
		Mutators:         cl.MutatorMap,
		Imports:          cl.Imports,
		AppInfo:          appInfo,
	}

	log.Printf("Generate package (%v)", cl)

	var err *arerror.ErrGeneratorPhases

	generated, err = generateFixture(params)
	if err != nil {
		err.Name = cl.Namespace.PublicName
		return nil, err
	}

	for _, data := range generated {
		genRes := GenerateFile{
			Dir:  pkgFixture,
			Name: cl.Namespace.PackageName + "_gen.go",
		}

		genData := data.Bytes()

		dataImp, err := imports.Process("", genData, nil)
		if err != nil {
			return nil, &arerror.ErrGeneratorFile{Name: cl.Namespace.PublicName, Backend: "fixture", Filename: genRes.Name, Err: ErrorLine(err, string(genData))}
		}

		genRes.Data = dataImp
		ret = append(ret, genRes)
	}

	return ret, nil
}
