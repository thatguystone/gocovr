package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/thatguystone/assert"
)

func testDump(t *testing.T, file, filter string) (assert.A, *bytes.Buffer, *bytes.Buffer) {
	a := assert.From(t)

	out := bytes.Buffer{}
	err := bytes.Buffer{}

	dump(&out, &err, file, filter)

	return a, &out, &err
}

func TestData0(t *testing.T) {
	a, out, err := testDump(t, "test_data/0", ".*")
	a.MustEqual(0, err.Len())

	tests := []string{
		`8.go\s*14\s*14\s*100.0%`,
		`5.go\s*63\s*51\s*81.0%`,
		`3.go\s*37\s*0\s*0.0%\s*31-108`,
		`19.go\s*1\s*0\s*0.0%\s*1`,
		`TOTAL\s*1533*\s*484\s*31.6%`,
	}

	for _, t := range tests {
		r := regexp.MustCompile(t)
		a.True(r.MatchString(out.String()), "%s did not match", t)
	}

	t.Log(out.String())

}

func TestData0Filter(t *testing.T) {
	a, out, _ := testDump(t, "test_data/0", "5.go")

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
	a, out, err := testDump(t, "test_data/1", ".*")

	a.Equal(0, out.Len())
	a.Equal("No files covered.\n", err.String())
}

func TestInvalidFilter(t *testing.T) {
	a, out, err := testDump(t, "test_data/1", "*")

	a.Equal(0, out.Len())
	a.True(strings.HasPrefix(err.String(), "Error: invalid filter:"),
		"Got string: %s", err.String())
}

func TestInvalidCoverageFile(t *testing.T) {
	a, out, err := testDump(t, "main.go", ".*")

	a.Equal(0, out.Len())
	a.True(strings.HasPrefix(err.String(), "Error: invalid coverage profile:"),
		"Got string: %s", err.String())
}

func TestData2(t *testing.T) {
	a, out, err := testDump(t, "test_data/2", ".*")
	a.MustEqual(0, err.Len())

	tests := []string{
		`0.go\s*0\s*0\s*100.0%`,
		`TOTAL\s*0*\s*0\s*100.0%`,
	}

	for _, t := range tests {
		r := regexp.MustCompile(t)
		a.True(r.MatchString(out.String()), "%s did not match", t)
	}
}
