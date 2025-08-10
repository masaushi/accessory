package accessor

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

// ParsedSource contains the parsed source of a package.
type ParsedSource struct {
	Package *packages.Package
	Dir     string
	Imports []*Import
	Structs []*Struct
}

// Import contains the information of an import.
type Import struct {
	Name    string
	Path    string
	IsNamed bool
}

// Struct contains the information of a struct.
type Struct struct {
	Name     string
	Fields   []*Field
	LockType LockType
}

// Field contains the information of a field in a struct.
type Field struct {
	Name string
	Type types.Type
	Tag  *Tag
}

// Tag contains the information of a struct field's tag.
type Tag struct {
	Getter    *string
	Setter    *string
	NoDefault bool
}

// LockType represents the type of lock used in a struct.
type LockType string

const (
	LockTypeNone    LockType = "none"
	LockTypeMutex   LockType = "mutex"
	LockTypeRWMutex LockType = "rwmutex"
)
