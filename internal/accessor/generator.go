package accessor

import (
	"bytes"
	"fmt"
	"go/types"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/afero"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
)

type generator struct {
	writer   *writer
	typ      string
	output   string
	receiver string
	lock     string
}

type methodGenParameters struct {
	Receiver     string
	Struct       string
	Field        string
	GetterMethod string
	SetterMethod string
	Type         string
	ZeroValue    string // used only when generating getter
	Lock         string
}

func newGenerator(fs afero.Fs, pkg *Package, options ...Option) *generator {
	g := new(generator)
	for _, opt := range options {
		opt(g)
	}

	path := g.outputFilePath(pkg.Dir)
	g.writer = newWriter(fs, path)

	return g
}

// Generate generates a file and accessor methods.
func Generate(fs afero.Fs, pkg *Package, options ...Option) error {
	g := newGenerator(fs, pkg, options...)

	importMap := make(map[string]*packages.Package, len(pkg.Imports))
	for _, imp := range pkg.Imports {
		// temporary assign nil
		importMap[imp.Name] = imp
	}

	accessors := make([]string, 0)
	imports := make([]*packages.Package, 0, len(importMap))

	for _, st := range pkg.Structs {
		if st.Name != g.typ {
			continue
		}

		for _, field := range st.Fields {
			if field.Tag == nil {
				continue
			}

			params := g.setupParameters(pkg, st, field)

			if field.Tag.Getter != nil {
				getter, err := g.generateGetter(params)
				if err != nil {
					return err
				}
				accessors = append(accessors, getter)
			}
			if field.Tag.Setter != nil {
				setter, err := g.generateSetter(params)
				if err != nil {
					return err
				}
				accessors = append(accessors, setter)
			}

			if splitted := strings.Split(strings.TrimPrefix(params.Type, "*"), "."); len(splitted) > 1 {
				otherPackage := splitted[0]
				imports = append(imports, importMap[otherPackage])
			}
		}
	}

	return g.writer.write(pkg.Name, g.generateImportStrings(imports), accessors)
}

func (g *generator) outputFilePath(dir string) string {
	output := g.output
	if output == "" {
		// Use snake_case name of type as output file if output file is not specified.
		// type TestStruct will be test_struct_accessor.go
		var firstCapMatcher = regexp.MustCompile("(.)([A-Z][a-z]+)")
		var articleCapMatcher = regexp.MustCompile("([a-z0-9])([A-Z])")

		name := firstCapMatcher.ReplaceAllString(g.typ, "${1}_${2}")
		name = articleCapMatcher.ReplaceAllString(name, "${1}_${2}")
		output = strings.ToLower(fmt.Sprintf("%s_accessor.go", name))
	}

	return filepath.Join(dir, output)
}

func (g *generator) generateSetter(
	params *methodGenParameters,
) (string, error) {
	var lockingCode string
	if params.Lock != "" {
		lockingCode = ` {{.Receiver}}.{{.Lock}}.Lock()
		defer {{.Receiver}}.{{.Lock}}.Unlock()
		`
	}

	var tpl = `
	func ({{.Receiver}} *{{.Struct}}) {{.SetterMethod}}(val {{.Type}}) {
		if {{.Receiver}} == nil {
			return
		}
	` +
		lockingCode + // inject locing code
		`{{.Receiver}}.{{.Field}} = val
	}`

	t := template.Must(template.New("setter").Parse(tpl))
	buf := new(bytes.Buffer)

	if err := t.Execute(buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (g *generator) generateGetter(
	params *methodGenParameters,
) (string, error) {
	var lockingCode string
	if params.Lock != "" {
		lockingCode = `{{.Receiver}}.{{.Lock}}.Lock()
		defer {{.Receiver}}.{{.Lock}}.Unlock()
		`
	}

	var tpl = `
	func ({{.Receiver}} *{{.Struct}}) {{.GetterMethod}}() {{.Type}} {
		if {{.Receiver}} == nil {
			return {{.ZeroValue}}
		}
		` +
		lockingCode + // inject locing code
		`return {{.Receiver}}.{{.Field}}
	}`

	t := template.Must(template.New("getter").Parse(tpl))
	buf := new(bytes.Buffer)

	if err := t.Execute(buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (g *generator) setupParameters(
	pkg *Package,
	st *Struct,
	field *Field,
) *methodGenParameters {
	typeName := g.typeName(pkg.Types, field.Type)
	getter, setter := g.methodNames(field)
	return &methodGenParameters{
		Receiver:     g.receiverName(st.Name),
		Struct:       st.Name,
		Field:        field.Name,
		GetterMethod: getter,
		SetterMethod: setter,
		Type:         typeName,
		ZeroValue:    g.zeroValue(field.Type, typeName),
		Lock:         g.lock,
	}
}

func (g *generator) receiverName(structName string) string {
	if g.receiver != "" {
		// Do nothing if receiver name specified in args.
		return g.receiver
	}

	// Use the first letter of struct as receiver if receiver name is not specified.
	return strings.ToLower(string(structName[0]))
}

func (g *generator) methodNames(field *Field) (getter, setter string) {
	if getterName := field.Tag.Getter; getterName != nil && *getterName != "" {
		getter = *getterName
	} else {
		getter = cases.Title(language.Und).String(field.Name)
	}

	if setterName := field.Tag.Setter; setterName != nil && *setterName != "" {
		setter = *setterName
	} else {
		setter = "Set" + cases.Title(language.Und).String(field.Name)
	}

	return getter, setter
}

func (g *generator) typeName(pkg *types.Package, t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		// type is defined in same package
		if pkg == p {
			return ""
		}
		// path string(like example.com/user/project/package) into slice
		return p.Name()
	})
}

func (g *generator) zeroValue(t types.Type, typeString string) string {
	switch t := t.(type) {
	case *types.Pointer:
		return "nil"
	case *types.Array:
		return "nil"
	case *types.Slice:
		return "nil"
	case *types.Chan:
		return "nil"
	case *types.Interface:
		return "nil"
	case *types.Map:
		return "nil"
	case *types.Signature:
		return "nil"
	case *types.Struct:
		return typeString + "{}"
	case *types.Basic:
		info := types.Typ[t.Kind()].Info()
		switch {
		case types.IsNumeric&info != 0:
			return "0"
		case types.IsBoolean&info != 0:
			return "false"
		case types.IsString&info != 0:
			return `""`
		}
	case *types.Named:
		if types.Identical(t, types.Universe.Lookup("error").Type()) {
			return "nil"
		}

		return g.zeroValue(t.Underlying(), typeString)
	}

	return "nil"
}

func (g *generator) generateImportStrings(pkgs []*packages.Package) []string {
	// Ensure imports are same order as previous if there are no declaration changes.
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Name < pkgs[j].Name
	})

	imports := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		if pkg.Name == filepath.Base(pkg.PkgPath) {
			imports[i] = pkg.PkgPath
		} else {
			imports[i] = fmt.Sprintf("%s \"%s\"", pkg.Name, pkg.PkgPath)
		}
	}

	return imports
}
