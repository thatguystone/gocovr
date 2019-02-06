package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

func testDump(file string, showCovered bool) (string, error) {
	var out bytes.Buffer
	err := dump(&out, file, showCovered)

	return out.String(), err
}

func hasLines(c *check.C, out string, lines []string) {
	for _, line := range lines {
		linere := strings.ReplaceAll(line, "\t", "\\s*")
		r := regexp.MustCompile(linere)
		c.Truef(r.MatchString(out), "%s did not match", line)
	}
}

func TestDumpBasic(t *testing.T) {
	c := check.New(t)

	out, err := testDump("testdata/basic/cover.out", false)
	c.Must.Nil(err)
	c.Log(out)

	c.NotContains(out, "a.go")
	hasLines(c, out, []string{
		"b.go	8	0	0.0%	3-20",
		"TOTAL	17	9	52.9%",
	})

}

func TestDumpShowCovered(t *testing.T) {
	c := check.New(t)

	out, err := testDump("testdata/basic/cover.out", true)
	c.Must.Nil(err)
	c.Log(out)

	hasLines(c, out, []string{
		"a.go	8	8	100.0%",
		"b.go	8	0	0.0%	3-20",
		"TOTAL	17	9	52.9%",
	})
}

func TestDumpNoFiles(t *testing.T) {
	c := check.New(t)

	out, err := testDump("testdata/nofiles/cover.out", true)
	c.Must.Nil(err)
	c.Equal(out, "No files covered.\n")
}

func TestDumpCompleteCoverage(t *testing.T) {
	c := check.New(t)

	out, err := testDump("testdata/fullcoverage/cover.out", false)
	c.Must.Nil(err)
	c.Log(out)

	c.NotContains(out, "a.go")
	hasLines(c, out, []string{
		"TOTAL	1	1	100.0%",
	})
}

func TestDumpCompleteCoverageShowCovered(t *testing.T) {
	c := check.New(t)

	out, err := testDump("testdata/fullcoverage/cover.out", true)
	c.Must.Nil(err)
	c.Log(out)

	hasLines(c, out, []string{
		"a.go	1	1	100.0%",
		"TOTAL	1	1	100.0%",
	})
}

func TestDumpInvalidCoverageFile(t *testing.T) {
	c := check.New(t)

	_, err := testDump("testdata/basic/a.go", true)
	c.NotNil(err)
}
