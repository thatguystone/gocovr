package testdata

func aBool(b bool) {
	if b {
		_ = "test"
	}

	if b {
		_ = "test"
	}

	if b {
		_ = "test"
	}
}

func aLoop(i int) {
	for j := 0; j < i; j++ {
		_ = "test"
	}
}
