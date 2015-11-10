package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

func testDump(t *testing.T, includeRe, excludeRe string, files ...string) (
	*check.C,
	*bytes.Buffer,
	[]error) {

	c := check.New(t)
	out := bytes.Buffer{}
	errs := dump(&out, files, includeRe, excludeRe)

	return c, &out, errs
}

func TestData0(t *testing.T) {
	c, out, errs := testDump(t, ".*", "^$", "test_fixtures/0", "test_fixtures/0")
	c.MustEqual(0, len(errs))

	tests := []string{
		`8.go\s*14\s*14\s*100.0%`,
		`5.go\s*63\s*51\s*81.0%`,
		`3.go\s*37\s*0\s*0.0%\s*31-108`,
		`19.go\s*1\s*0\s*0.0%\s*1`,
		`TOTAL\s*1459\s*431\s*29.5%`,
	}

	c.Log(out)

	for _, t := range tests {
		r := regexp.MustCompile(t)
		c.True(r.MatchString(out.String()), "%s did not match", t)
	}
}

func TestData0Filter(t *testing.T) {
	c, out, _ := testDump(t, "5.go", "^$", "test_fixtures/0")

	tests := []string{
		`5.go\s*63\s*51\s*81.0%`,
		`TOTAL\s*128*\s*51\s*39.8%`,
	}

	c.Log(out)

	for _, t := range tests {
		r := regexp.MustCompile(t)
		c.True(r.MatchString(out.String()), "%s did not match", t)
	}
}

func TestData0Exclude(t *testing.T) {
	c, out, _ := testDump(t, ".*", `[0-46-9]\.go`, "test_fixtures/0")

	tests := []string{
		`5.go\s*63\s*51\s*81.0%`,
		`TOTAL\s*128*\s*51\s*39.8%`,
	}

	c.Log(out)

	for _, t := range tests {
		r := regexp.MustCompile(t)
		c.True(r.MatchString(out.String()), "%s did not match", t)
	}
}

func TestData1(t *testing.T) {
	c, out, errs := testDump(t, ".*", "^$", "test_fixtures/1")

	c.Equal(0, len(errs))
	c.Equal("No files covered.\n", out.String())
}

func TestInvalidFilter(t *testing.T) {
	c, out, errs := testDump(t, "*", "^$", "test_fixtures/1")
	c.Equal(0, out.Len())
	c.True(strings.HasPrefix(errs[0].Error(), "invalid include pattern:"),
		"Got error: %s", errs[0].Error())

	c, out, errs = testDump(t, ".*", "*", "test_fixtures/1")
	c.Equal(0, out.Len())
	c.True(strings.HasPrefix(errs[0].Error(), "invalid exclude pattern:"),
		"Got error: %s", errs[0].Error())
}

func TestInvalidCoverageFile(t *testing.T) {
	c, out, errs := testDump(t, ".*", "^$", "main.go")

	c.Log(out)

	c.Equal(0, out.Len())
	c.True(strings.HasPrefix(errs[0].Error(), "invalid coverage profile:"),
		"Got string: %s", errs[0].Error())
}

func TestData2(t *testing.T) {
	c, out, errs := testDump(t, ".*", "^$", "test_fixtures/2")
	c.MustEqual(0, len(errs))

	tests := []string{
		`0.go\s*0\s*0\s*100.0%`,
		`TOTAL\s*0*\s*0\s*100.0%`,
	}

	c.Log(out)

	for _, t := range tests {
		r := regexp.MustCompile(t)
		c.True(r.MatchString(out.String()), "%s did not match", t)
	}
}
