package main

func lcp(a, b string) string {
	if len(b) < len(a) {
		a, b = b, a
	}

	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}

	return a
}
