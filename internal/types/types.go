package types

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	Dir   string
	Name  string
	Files []*File
}

type File struct {
	File    *ast.File
	Imports map[string]*packages.Package
	Structs []*Struct
}

type Struct struct {
	Name   string
	Fields []*Field
}

type Field struct {
	Name     string
	DataType string
	Tag      *Tag
}

type Tag struct {
	Getter *string
	Setter *string
}
