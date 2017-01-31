package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"golang.org/x/tools/cover"
)

const (
	fileSkipDirective = "//gocovr:skip-file"
)

func dump(outW io.Writer, files []string, includeRe, excludeRe string, showCovered bool) (errs []error) {
	includePat, err := regexp.Compile(includeRe)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid include pattern: %s", err))
		return
	}

	excludePat, err := regexp.Compile(excludeRe)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid exclude pattern: %s", err))
		return
	}

	profs := profiles{}
	pll := parallelize{}
	pll.do(files, func(file string) error {
		ps, err := cover.ParseProfiles(file)
		if err != nil {
			return fmt.Errorf("invalid coverage profile: %s", err)
		}

		for _, p := range ps {
			ignore := ignoreFile(p.FileName) ||
				!includePat.Match([]byte(p.FileName)) ||
				excludePat.Match([]byte(p.FileName))
			if ignore {
				continue
			}

			profs.add(p)
		}

		return nil
	})

	pll.Wait()
	errs = pll.errs

	if len(errs) > 0 {
		return
	}

	if len(profs.m) == 0 {
		fmt.Fprintln(outW, "No files covered.")
		return
	}

	base := profs.getBase()
	fmt.Fprintf(outW, "Base: %s\n\n", profs.getBase())

	w := tabwriter.NewWriter(outW, 0, 4, 2, ' ', 0)
	defer w.Flush()

	print(w, "File", "Lines", "Exec", "Cover", "Missing")
	printLine(w)

	totalLines := 0
	totalExec := 0

	sorted := profs.sortedSlice()
	for _, p := range sorted {
		lines := 0
		exec := 0
		missing := []string{}

		for _, b := range coalesce(p.Blocks) {
			lines += b.NumStmt
			if b.Count > 0 {
				exec += b.NumStmt
			}

			if b.Count == 0 && b.NumStmt > 0 {
				if b.StartLine == b.EndLine {
					missing = append(missing,
						fmt.Sprintf("%d", b.StartLine))
				} else {
					missing = append(missing,
						fmt.Sprintf("%d-%d",
							b.StartLine,
							b.EndLine))
				}
			}
		}

		if len(missing) > 0 || showCovered {
			printSummary(w,
				strings.TrimPrefix(p.FileName, base),
				lines, exec,
				strings.Join(missing, ","))
		}

		totalLines += lines
		totalExec += exec
	}

	printLine(w)
	printSummary(w,
		"TOTAL",
		totalLines, totalExec,
		"")

	return
}

func coalesce(pbs []cover.ProfileBlock) []cover.ProfileBlock {
	ret := []cover.ProfileBlock{}

	join := func(a, b cover.ProfileBlock) bool {
		// Two "misses" next to each other can always be joined
		return a.Count == 0 && b.Count == 0
	}

	for i := 0; i < len(pbs); i++ {
		b := pbs[i]
		pb := b

		for (i+1) < len(pbs) && join(pb, pbs[i+1]) {
			npb := pbs[i+1]
			pb.EndLine = npb.EndLine
			pb.EndCol = npb.EndCol
			pb.NumStmt += npb.NumStmt
			pb.Count += npb.Count
			i++
		}

		ret = append(ret, pb)
	}

	return ret
}

func print(w *tabwriter.Writer, file, lines, exec, cover, missing string) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", file, lines, exec, cover, missing)
}

func printLine(w *tabwriter.Writer) {
	print(w, "", "", "", "", "")
}

func printSummary(w *tabwriter.Writer, name string, lines, exec int, missing string) {
	covered := float64(exec) / float64(lines)
	if exec == 0 && lines == 0 {
		covered = 1
	}

	print(w,
		name,
		fmt.Sprintf("%d", lines),
		fmt.Sprintf("%d", exec),
		fmt.Sprintf("%0.1f%%", covered*100),
		missing)
}

func findFile(fileName string) *os.File {
	for _, dir := range strings.Split(os.Getenv("GOPATH"), ":") {
		abspath := fmt.Sprintf("%s/src/%s", dir, fileName)
		f, err := os.Open(abspath)
		if err == nil {
			return f
		}
	}

	return nil
}

func ignoreFile(fileName string) bool {
	f := findFile(fileName)
	if f == nil {
		return false
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == fileSkipDirective {
			return true
		}
	}

	return false
}
