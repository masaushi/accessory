package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/masaushi/accessory/internal/types"

	"golang.org/x/tools/go/packages"
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
func ParsePackage(dir string) (*types.Package, error) {
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

	return &types.Package{Dir: dir, Name: pkgs[0].Name, Files: parseFiles(pkgs[0])}, nil
}

func parseFiles(pkg *packages.Package) []*types.File {
	files := make([]*types.File, len(pkg.Syntax))
	for i := range pkg.Syntax {
		files[i] = &types.File{
			File:    pkg.Syntax[i],
			Imports: pkg.Imports,
			Structs: parseStructs(pkg.Fset, pkg.Syntax[i]),
		}
	}

	return files
}

func parseStructs(fileSet *token.FileSet, file *ast.File) []*types.Struct {
	structs := make([]*types.Struct, 0)

	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Type == nil {
			return true
		}

		s, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structs = append(structs, &types.Struct{
			Name:   ts.Name.Name,
			Fields: parseFields(s, fileSet),
		})

		return false
	})

	return structs
}

func parseFields(st *ast.StructType, fileSet *token.FileSet) []*types.Field {
	fields := make([]*types.Field, 0)
	for _, field := range st.Fields.List {
		if field.Tag == nil {
			continue
		}

		name := field.Names[0].Name
		buf := new(bytes.Buffer)
		printer.Fprint(buf, fileSet, field.Type)
		sf := &types.Field{
			Name:     name,
			DataType: buf.String(),
			Tag:      parseTag(field.Tag),
		}
		fields = append(fields, sf)
	}

	return fields
}

func parseTag(tag *ast.BasicLit) *types.Tag {
	tagStr, ok := reflect.StructTag(strings.Trim(tag.Value, "`")).Lookup(accessorTag)
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

	return &types.Tag{Setter: setter, Getter: getter}
}
