package testdata

func bBool(b bool) {
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

func bLoop(i int) {
	for j := 0; j < i; j++ {
		_ = "test"
	}
}
