package test

import "sync"

type Tester struct {
	lock   sync.Cond
	field1 string `accessor:"getter:GetField1,setter"`
	field2 int32  `accessor:"getter:GetField2,setter"`
	field3 *bool
}
