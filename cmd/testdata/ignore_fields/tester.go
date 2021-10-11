package test

type Tester struct {
	field1 string `accessor:"-"`
	field2 int32  `accessor:"getter"`
	field3 *bool
}
