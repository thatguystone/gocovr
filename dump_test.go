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

func hasLines(t *testing.T, out string, lines []string) {
	for _, line := range lines {
		linere := strings.ReplaceAll(line, "\t", "\\s*")
		r := regexp.MustCompile(linere)
		check.Truef(t, r.MatchString(out), "%s did not match", line)
	}
}

func TestDumpBasic(t *testing.T) {
	t.Parallel()

	out, err := testDump("testdata/basic/cover.out", false)
	check.MustNil(t, err)
	t.Log(out)

	check.False(t, strings.Contains(out, "a.go"))
	hasLines(t, out, []string{
		"b.go	8	0	0.0%	3-20",
		"TOTAL	17	9	52.9%",
	})

}

func TestDumpShowCovered(t *testing.T) {
	t.Parallel()

	out, err := testDump("testdata/basic/cover.out", true)
	check.MustNil(t, err)
	t.Log(out)

	hasLines(t, out, []string{
		"a.go	8	8	100.0%",
		"b.go	8	0	0.0%	3-20",
		"TOTAL	17	9	52.9%",
	})
}

func TestDumpNoFiles(t *testing.T) {
	t.Parallel()

	out, err := testDump("testdata/nofiles/cover.out", true)
	check.MustNil(t, err)
	check.Equal(t, out, "No files covered.\n")
}

func TestDumpCompleteCoverage(t *testing.T) {
	t.Parallel()

	out, err := testDump("testdata/fullcoverage/cover.out", false)
	check.MustNil(t, err)
	t.Log(out)

	check.False(t, strings.Contains(out, "a.go"))
	hasLines(t, out, []string{
		"TOTAL	1	1	100.0%",
	})
}

func TestDumpCompleteCoverageShowCovered(t *testing.T) {
	t.Parallel()

	out, err := testDump("testdata/fullcoverage/cover.out", true)
	check.MustNil(t, err)
	t.Log(out)

	hasLines(t, out, []string{
		"a.go	1	1	100.0%",
		"TOTAL	1	1	100.0%",
	})
}

func TestDumpInvalidCoverageFile(t *testing.T) {
	t.Parallel()

	_, err := testDump("testdata/basic/a.go", true)
	check.NotNil(t, err)
}
