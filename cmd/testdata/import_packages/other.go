package test

import (
	"github.com/masaushi/accessory/cmd/testdata/import_packages/sub2"
)

type Other struct {
	field4 *sub2.SubTester `accessor:"getter,setter"`
}
