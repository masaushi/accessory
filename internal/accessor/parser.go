package accessor

import (
	"fmt"
	"go/types"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	accessorTag     = "accessor"
	ignoreTag       = "-"
	tagKeyGetter    = "getter"
	tagKeySetter    = "setter"
	tagKeyNoDefault = "noDefault"
)

const (
	tagSep         = ","
	tagKeyValueSep = ":"
)

func Parse(dir string) (*ParsedSource, error) {
	const mode = packages.NeedName | packages.NeedFiles |
		packages.NeedImports | packages.NeedTypes | packages.NeedSyntax

	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode:  mode,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("error: %d packages found", len(pkgs))
	}

	return &ParsedSource{
		Package: pkgs[0],
		Dir:     dir,
		Imports: parseImports(pkgs[0]),
		Structs: parseStructs(pkgs[0]),
	}, nil
}

func parseImports(pkg *packages.Package) []*Import {
	var imports []*Import

	for _, syntax := range pkg.Syntax {
		for _, imp := range syntax.Imports {
			// Extract the path from the import. Remove the leading and trailing quotes.
			path := strings.Trim(imp.Path.Value, "\"")

			// Extract the name from the import. If the import is not named, use the base name of the path.
			name := filepath.Base(path)
			isNamed := false
			if imp.Name != nil {
				name = imp.Name.Name
				isNamed = true
			}
			if !slices.ContainsFunc(imports, func(imp *Import) bool {
				return imp.Name == name &&
					imp.Path == path &&
					imp.IsNamed == isNamed
			}) {
				imports = append(imports, &Import{
					Name:    name,
					Path:    path,
					IsNamed: isNamed,
				})
			}
		}
	}

	return imports
}

func parseStructs(pkg *packages.Package) []*Struct {
	scope := pkg.Types.Scope()
	structs := make([]*Struct, 0, len(scope.Names()))
	for _, name := range scope.Names() {
		st, ok := scope.Lookup(name).Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}

		structs = append(structs, &Struct{
			Name:   name,
			Fields: parseFields(st),
		})
	}

	return structs
}

func parseFields(st *types.Struct) []*Field {
	fields := make([]*Field, st.NumFields())
	for i := 0; i < st.NumFields(); i++ {
		tag := parseTag(st.Tag(i))
		field := st.Field(i)

		fields[i] = &Field{
			Name: field.Name(),
			Type: field.Type(),
			Tag:  tag,
		}
	}

	return fields
}

func parseTag(tag string) *Tag {
	tagStr, ok := reflect.StructTag(strings.Trim(tag, "`")).Lookup(accessorTag)
	if !ok {
		return nil
	}

	var getter, setter *string
	var noDefault bool

	tags := strings.Split(tagStr, tagSep)
	for _, tag := range tags {
		keyValue := strings.Split(tag, tagKeyValueSep)

		var value string
		if len(keyValue) == 2 {
			if v := strings.TrimSpace(keyValue[1]); v != ignoreTag {
				value = v
			}
		}
		switch strings.TrimSpace(keyValue[0]) {
		case tagKeyGetter:
			getter = &value
		case tagKeySetter:
			setter = &value
		case tagKeyNoDefault:
			noDefault = true
		}
	}

	return &Tag{Setter: setter, Getter: getter, NoDefault: noDefault}
}
