package test

type Tester struct {
	field1 string `accessor:"getter,noDefault"`
	field2 int32  `accessor:"getter,setter,noDefault"`
	field3 *bool  `accessor:"getter:GetFieldThree,noDefault"`
}
