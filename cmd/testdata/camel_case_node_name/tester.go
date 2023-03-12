package test

type Tester struct {
	firstField  string `accessor:"getter"`
	secondField int32  `accessor:"setter"`
	thirdField  int32  `accessor:"getter,setter"`
}
