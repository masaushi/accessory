package test

type Tester struct {
	field1 string `accessor:"getter"`
	field2 int32  `accessor:"setter:SetSecondField"`
	field3 *bool
}
