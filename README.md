# accessory

[![test](https://github.com/masaushi/accessory/actions/workflows/test.yml/badge.svg)](https://github.com/masaushi/accessory/actions/workflows/test.yml)
[![release](https://github.com/masaushi/accessory/actions/workflows/release.yml/badge.svg)](https://github.com/masaushi/accessory/actions/workflows/release.yml)

accessory is an accessor generator for [Go programming language](https://golang.org/).

## What is accessory?

Accessory is a tool that generates accessor methods from any structs.

Sometimes you might make struct fields unexported in order for values of fields not to be accessed
or modified from anywhere in your codebases, and define getters or setters for values to be handled in a desired way.

But writing accessors for so many fields is time-consuming, but not exciting or creative.

Accessory frees you from tedious, monotonous tasks.

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
    return m.field1
}

func(m *MyStruct) SetField2(val *int) {
    m.field2 = val
}

func(m *MyStruct) Field3() time.Time {
    return m.field3
}

func(m *MyStruct) SetField3(val time.Time) {
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
    return m.field1
}

func(m *MyStruct) ChangeSecondField(val *int) {
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

  -receiver string
      receiver receiver for generated accessor methods
      default: first letter of struct

  -output string
      output file name
      default: <type_name>_accessor.go

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
