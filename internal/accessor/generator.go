package accessor

import (
	"bytes"
	"fmt"
	"go/types"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	templates "github.com/masaushi/accessory/internal/accessor/gotemplates"
	"github.com/spf13/afero"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
)

type generator struct {
	writer *writer

	typ      string
	output   string
	receiver string
	lock     string

	pkg     *packages.Package
	imports []*Import

	usedPackages map[string]struct{}
}

type methodGenParameters struct {
	Receiver     string
	Struct       string
	Field        string
	GetterMethod string
	SetterMethod string
	NoDefault    bool
	Type         string
	ZeroValue    string // used only when generating getter
	Lock         string
}

func newGenerator(fs afero.Fs, src *ParsedSource, options ...Option) *generator {
	g := new(generator)
	for _, opt := range options {
		opt(g)
	}

	path := g.outputFilePath(src.Dir)
	g.writer = newWriter(fs, path)

	g.pkg = src.Package
	g.imports = src.Imports
	g.usedPackages = make(map[string]struct{})

	return g
}

func Generate(fs afero.Fs, src *ParsedSource, options ...Option) error {
	g := newGenerator(fs, src, options...)

	// Generate accessor methods for the specified type.
	accessors, err := g.generateAccessors(src.Structs)
	if err != nil {
		return err
	}

	// Generate import statements for used packages.
	imports := g.generateImports()

	// Write the generated content to the file system.
	return g.writer.write(src.Package.Name, imports, accessors)
}

func (g *generator) outputFilePath(dir string) string {
	output := g.output
	// If output file path is not specified, use snake_case name of the type as output file.
	if output == "" {
		// Convert the first letter of the type to lowercase and replace all uppercase letters
		// followed by lowercase letters with the lowercase letter preceded by an underscore.
		// For example, "TestStruct" becomes "test_struct".
		var firstCapMatcher = regexp.MustCompile("(.)([A-Z][a-z]+)")
		var articleCapMatcher = regexp.MustCompile("([a-z0-9])([A-Z])")

		name := firstCapMatcher.ReplaceAllString(g.typ, "${1}_${2}")
		name = articleCapMatcher.ReplaceAllString(name, "${1}_${2}")
		output = strings.ToLower(fmt.Sprintf("%s_accessor.go", name))
	}

	return filepath.Join(dir, output)
}

func (g *generator) generateImports() []string {
	importStrings := make([]string, 0, len(g.imports))

	for _, imp := range g.imports {
		if _, ok := g.usedPackages[imp.Path]; !ok {
			continue
		}

		importString := fmt.Sprintf("%q", imp.Path)
		if imp.IsNamed {
			// If the import is named, add the name before the path.
			importString = imp.Name + " " + importString
		}

		importStrings = append(importStrings, importString)
	}

	return importStrings
}

func (g *generator) generateAccessors(structs []*Struct) ([]string, error) {
	accessors := make([]string, 0)

	for _, st := range structs {
		// Check if the struct name matches the type name of the generator.
		if st.Name != g.typ {
			continue
		}

		for _, field := range st.Fields {
			if field.Tag == nil {
				continue
			}

			params := g.createMethodGenParameters(st, field)

			if field.Tag.Getter != nil {
				getter, err := g.generateGetter(params)
				if err != nil {
					return nil, err
				}
				accessors = append(accessors, getter)
			}
			if field.Tag.Setter != nil {
				setter, err := g.generateSetter(params)
				if err != nil {
					return nil, err
				}
				accessors = append(accessors, setter)
			}

			if usedPackage := g.getUsedPackages(field); usedPackage != "" {
				g.usedPackages[usedPackage] = struct{}{}
			}
		}
	}

	return accessors, nil
}

func (g *generator) generateSetter(
	params *methodGenParameters,
) (string, error) {
	t := template.Must(template.New("setter").Parse(templates.Setter))
	buf := new(bytes.Buffer)

	if err := t.Execute(buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (g *generator) generateGetter(
	params *methodGenParameters,
) (string, error) {
	// Template
	var tmpl = templates.Getter
	if params.NoDefault {
		tmpl = templates.GetterNoDefault
	}

	t := template.Must(template.New("getter").Parse(tmpl))
	buf := new(bytes.Buffer)

	if err := t.Execute(buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (g *generator) createMethodGenParameters(st *Struct, field *Field) *methodGenParameters {
	typeName := g.typeName(field.Type)
	getter, setter := g.methodNames(field)
	return &methodGenParameters{
		Receiver:     g.receiverName(st.Name),
		Struct:       st.Name,
		Field:        field.Name,
		GetterMethod: getter,
		SetterMethod: setter,
		NoDefault:    field.Tag.NoDefault,
		Type:         typeName,
		ZeroValue:    g.zeroValue(field.Type, typeName),
		Lock:         g.lock,
	}
}

func (g *generator) getUsedPackages(field *Field) string {
	var typePackage string
	types.TypeString(field.Type, func(p *types.Package) string {
		typePackage = p.Path()
		return ""
	})

	return typePackage
}

func (g *generator) receiverName(structName string) string {
	// If a receiver name is specified in the arguments, use it.
	if g.receiver != "" {
		return g.receiver
	}

	// If no receiver name is specified, use the first letter of the struct name as receiver.
	return strings.ToLower(string(structName[0]))
}

func (g *generator) methodNames(field *Field) (getter, setter string) {
	if getterName := field.Tag.Getter; getterName != nil && *getterName != "" {
		getter = *getterName
	} else {
		// If no getter name is specified in the tag,
		// use the field name capitalized as the getter name.
		getter = cases.Title(language.Und, cases.NoLower).String(field.Name)
	}

	if setterName := field.Tag.Setter; setterName != nil && *setterName != "" {
		setter = *setterName
	} else {
		// If no setter name is specified in the tag,
		// use "Set" concatenated with the field name capitalized as the setter name.
		setter = "Set" + cases.Title(language.Und, cases.NoLower).String(field.Name)
	}

	return getter, setter
}

func (g *generator) typeName(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		// type is defined in the same package
		if g.pkg.Types == p {
			return "" // return an empty string
		}

		idx := slices.IndexFunc(g.imports, func(imp *Import) bool {
			return imp.Path == p.Path()
		})

		// If the package is not in imports but is a valid package, use its name
		if idx == -1 {
			return p.Name()
		}

		// get the import statement for the package that the type is defined in
		imp := g.imports[idx]

		if imp.Name == "." {
			// return an empty string if the type is defined in the current package
			return ""
		}

		// If import has an alias, use it, otherwise use package name
		if imp.Name != "" {
			return imp.Name
		}
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
