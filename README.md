# accessory

[![lint and test](https://github.com/masaushi/accessory/actions/workflows/lint_and_test.yml/badge.svg)](https://github.com/masaushi/accessory/actions/workflows/lint_and_test.yml)
[![release](https://github.com/masaushi/accessory/actions/workflows/release.yml/badge.svg)](https://github.com/masaushi/accessory/actions/workflows/release.yml)

Accessory is a tool designed for the [Go programming language](https://golang.org/) that automates the generation of accessor methods for struct fields. It helps developers maintain encapsulation by allowing struct fields to be unexported, while still providing controlled access through generated getter and setter methods.

## Why Use Accessory?
- **Encapsulation**: Keep your struct fields unexported, ensuring they're not accessed or modified unexpectedly.
- **Automation**: Writing accessors for numerous fields can be repetitive. Accessory automates this task, saving you time.
- **Customization**: While the tool follows Go conventions by default, you can customize accessor names as needed.

## Installation

To get the latest released version

### Go version < 1.16

```bash
go get github.com/masaushi/accessory
```

### Go 1.16+

```bash
go install github.com/masaushi/accessory@latest
```

## Usage

### Declare Struct with `accessor` Tag

`accessory` generates accessor methods from defined structs, so you need to declare a struct and fields with `accessor` tag.

Values for `accessor` tag is `getter` and `setter`, `getter` is for generating getter method and `setter` is for setter methods.

Here is an example:

```go
type MyStruct struct {
    field1 string    `accessor:"getter"`
    field2 *int      `accessor:"setter"`
    field3 time.Time `accessor:"getter,setter"`
}
```

Generated methods will be
```go
func(m *MyStruct) Field1() string {
    if m == nil {
        return ""
    }
    return m.field1
}

func(m *MyStruct) SetField2(val *int) {
    if m == nil {
        return
    }
    m.field2 = val
}

func(m *MyStruct) Field3() time.Time {
    if m == nil {
        return time.Time{}
    }
    return m.field3
}

func(m *MyStruct) SetField3(val time.Time) {
    if m == nil {
        return
    }
    m.field3 = val
}
```

Following to [convention](https://golang.org/doc/effective_go#Getters),
setter's name is `Set<FieldName>()` and getter's name is `<FieldName>()` by default,
in other words, `Set` will be put into setter's name and `Get` will **not** be put into getter's name.

You can customize names for setter and getter if you want.

```go
type MyStruct struct {
    field1 string `accessor:"getter:GetFirstField"`
    field2 int    `accessor:"setter:ChangeSecondField"`
}
```

Generated methods will be

```go
func(m *MyStruct) GetFirstField() string {
    if m == nil {
        return ""
    }
    return m.field1
}

func(m *MyStruct) ChangeSecondField(val *int) {
    if m == nil {
        return
    }
    m.field2 = val
}
```

Accessor methods won't be generated if `accessor` tag isn't specified.
But you can explicitly skip generation by using `-` for tag value.

```go
type MyStruct struct {
    ignoredField `accessor:"-"`
}
```

### Run `accessory` command

To generate accessor methods, you need to run `accessory` command.

```
$ accessory [flags] source-dir

source-dir
  source-dir is the directory where the definition of the target struct is located.
  If source-dir is not specified, current directory is set as source-dir.

flags
  -type string <required>
      name of target struct

  -receiver string <optional>
      receiver receiver for generated accessor methods
      default: first letter of struct

  -output string <optional>
      output file name
      default: <type_name>_accessor.go

  -lock string <optional>
      specify lock field name and generate codes obtaining and releasing lock
      this is used to prevent race condition when concurrent access can be expected

  -version
      show the current version of accessory
```

Example:

```shell
$ accessory -type MyStruct -receiver myStruct -output my_struct_accessor.go path/to/target
```

#### go generate

You can also generate accessors by using `go generate`.

```go
package mypackage

//go:generate accessory -type MyStruct -receiver myStruct -output my_struct_accessor.go

type MyStruct struct {
    field1 string `accessor:"getter"`
    field2 *int   `accessor:"setter"`
}
```

Then run go generate for your package.

## License
The Accessory project (and all code) is licensed under the [MIT License](LICENSE).
