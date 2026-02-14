package testdata

import "fmt"

var assert asserter

type asserter interface {
	Unreachable()
}

func things(a int) {
	switch a {
	case 0:
		return
	case 1:
		panic("unreachable")
	case 2:
		panic("UNREACHABLE")
	case 3:
		assert.Unreachable()
	case 4:
		a = 1
		panic(fmt.Errorf("unreachable a=%v", a))
	case 5:
		// oh hey
		panic(fmt.Errorf(`unreachable a=%v`, a))
	case 6:
		// oh hey
		panic(fmt.Sprintf(`unreachable a=%v`, a))
	}
}
