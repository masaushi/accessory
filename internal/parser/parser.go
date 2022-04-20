package parser

import (
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/masaushi/accessory/internal/objects"
)

const (
	accessorTag  = "accessor"
	ignoreTag    = "-"
	tagKeyGetter = "getter"
	tagKeySetter = "setter"
)

const (
	tagSep         = ","
	tagKeyValueSep = ":"
)

// ParsePackage parses the specified directory's package.
func ParsePackage(dir string) (*objects.Package, error) {
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

	return &objects.Package{
		Package: pkgs[0],
		Dir:     dir,
		Structs: parseStructs(pkgs[0]),
	}, nil
}

func parseStructs(pkg *packages.Package) []*objects.Struct {
	scope := pkg.Types.Scope()
	structs := make([]*objects.Struct, 0, len(scope.Names()))
	for _, name := range scope.Names() {
		st, ok := scope.Lookup(name).Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}

		structs = append(structs, &objects.Struct{
			Name:   name,
			Fields: parseFields(pkg.Fset, st),
		})
	}

	return structs
}

func parseFields(fset *token.FileSet, st *types.Struct) []*objects.Field {
	fields := make([]*objects.Field, st.NumFields())
	for i := 0; i < st.NumFields(); i++ {
		tag := parseTag(st.Tag(i))
		field := st.Field(i)

		fields[i] = &objects.Field{
			Name: field.Name(),
			Type: field.Type(),
			Tag:  tag,
		}
	}

	return fields
}

func parseTag(tag string) *objects.Tag {
	tagStr, ok := reflect.StructTag(strings.Trim(tag, "`")).Lookup(accessorTag)
	if !ok {
		return nil
	}

	var getter, setter *string

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
		}
	}

	return &objects.Tag{Setter: setter, Getter: getter}
}
