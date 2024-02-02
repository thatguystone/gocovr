package main

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestLCP(t *testing.T) {
	tests := []struct {
		a   string
		b   string
		lcp string
	}{
		{
			a:   "test",
			b:   "tes",
			lcp: "tes",
		},
		{
			a:   "tes",
			b:   "test",
			lcp: "tes",
		},
		{
			a:   "test/stuff/",
			b:   "test/merp",
			lcp: "test/",
		},
	}

	for _, test := range tests {
		res := lcp(test.a, test.b)
		check.Equal(t, res, test.lcp)
	}
}
