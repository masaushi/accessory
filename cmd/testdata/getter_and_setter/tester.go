package test

type Tester struct {
	field1 string `accessor:"getter,setter"`
	field2 int32  `accessor:"getter:GetSecondField,setter:SetSecondField"`
	field3 *bool
	Field4 string `accessor:"getter,setter"`
	Field5 int32
}
