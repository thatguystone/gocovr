package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/thatguystone/assert"
)

func testDump(t *testing.T, filter string, files ...string) (
	assert.A,
	*bytes.Buffer,
	[]error) {

	t.Parallel()
	a := assert.From(t)

	out := bytes.Buffer{}

	errs := dump(&out, files, filter)

	return a, &out, errs
}

func TestData0(t *testing.T) {
	a, out, errs := testDump(t, ".*", "test_data/0", "test_data/0")
	a.MustEqual(0, len(errs))

	tests := []string{
		`8.go\s*14\s*14\s*100.0%`,
		`5.go\s*63\s*51\s*81.0%`,
		`3.go\s*37\s*0\s*0.0%\s*31-108`,
		`19.go\s*1\s*0\s*0.0%\s*1`,
		`TOTAL\s*1459\s*431\s*29.5%`,
	}

	for _, t := range tests {
		r := regexp.MustCompile(t)
		a.True(r.MatchString(out.String()), "%s did not match", t)
	}

	t.Log(out.String())
}

func TestData0Filter(t *testing.T) {
	a, out, _ := testDump(t, "5.go", "test_data/0")

	tests := []string{
		`5.go\s*63\s*51\s*81.0%`,
		`TOTAL\s*128*\s*51\s*39.8%`,
	}

	for _, t := range tests {
		r := regexp.MustCompile(t)
		a.True(r.MatchString(out.String()), "%s did not match", t)
	}
}

func TestData1(t *testing.T) {
	a, out, errs := testDump(t, ".*", "test_data/1")

	a.Equal(0, len(errs))
	a.Equal("No files covered.\n", out.String())
}

func TestInvalidFilter(t *testing.T) {
	a, out, errs := testDump(t, "*", "test_data/1")

	a.Equal(0, out.Len())
	a.True(strings.HasPrefix(errs[0].Error(), "invalid filter:"),
		"Got error: %s", errs[0].Error())
}

func TestInvalidCoverageFile(t *testing.T) {
	a, out, errs := testDump(t, ".*", "main.go")

	a.Equal(0, out.Len())
	a.True(strings.HasPrefix(errs[0].Error(), "invalid coverage profile:"),
		"Got string: %s", errs[0].Error())
}

func TestData2(t *testing.T) {
	a, out, errs := testDump(t, ".*", "test_data/2")
	a.MustEqual(0, len(errs))

	tests := []string{
		`0.go\s*0\s*0\s*100.0%`,
		`TOTAL\s*0*\s*0\s*100.0%`,
	}

	for _, t := range tests {
		r := regexp.MustCompile(t)
		a.True(r.MatchString(out.String()), "%s did not match", t)
	}
}
