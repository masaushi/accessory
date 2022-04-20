package objects

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	*packages.Package
	Dir     string
	Structs []*Struct
}

type Struct struct {
	Name   string
	Fields []*Field
}

type Field struct {
	Name string
	Type types.Type
	Tag  *Tag
}

type Tag struct {
	Getter *string
	Setter *string
}
