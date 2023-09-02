package test

import (
	"time"

	"github.com/masaushi/accessory/cmd/testdata/import_packages/sub1"
	sub "github.com/masaushi/accessory/cmd/testdata/import_packages/sub2"
	. "github.com/masaushi/accessory/cmd/testdata/import_packages/sub3"
	_ "github.com/masaushi/accessory/cmd/testdata/import_packages/sub4"
)

type Tester struct {
	field1 time.Time        `accessor:"getter,setter"`
	field2 *time.Time       `accessor:"getter,setter"`
	field3 *sub1.SubTester  `accessor:"getter,setter"`
	field4 *sub.SubTester   `accessor:"getter,setter"`
	field5 *SubTester       `accessor:"getter,setter"`
	field6 []sub.SubTester  `accessor:"getter,setter"`
	field7 []*sub.SubTester `accessor:"getter,setter"`
}
