package foo

import "time"

type Tester struct {
	time time.Time `accessor:"getter,setter"`
}
