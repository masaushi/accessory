package test

import (
	"time"

	sub "github.com/masaushi/accessory/cmd/testdata/import_packages/sub_package"
)

type Tester struct {
	field1 time.Time      `accessor:"getter"`
	field2 *sub.SubTester `accessor:"setter"`
	field3 *bool
}
