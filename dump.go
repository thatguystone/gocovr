package main

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"text/tabwriter"

	"golang.org/x/tools/cover"
)

func dump(outW, errW io.Writer, file, filter string) {
	log.SetFlags(0)
	log.SetOutput(errW)

	pat, err := regexp.Compile(filter)
	if err != nil {
		log.Printf("Error: invalid filter: %s", err)
		return
	}

	profs, err := cover.ParseProfiles(file)
	if err != nil {
		log.Printf("Error: invalid coverage profile: %s", err)
		return
	}

	if len(profs) == 0 {
		fmt.Fprintln(errW, "No files covered.")
		return
	}

	base := profs[0].FileName
	for _, p := range profs[1:] {
		base = lcp(base, p.FileName)
	}
	fmt.Fprintf(outW, "Base: %s\n\n", base)

	w := tabwriter.NewWriter(outW, 0, 4, 2, ' ', 0)
	defer w.Flush()

	print(w, "File", "Lines", "Exec", "Cover", "Missing")
	printLine(w)

	totalLines := 0
	totalExec := 0

	for _, p := range profs {
		if !pat.Match([]byte(p.FileName)) {
			continue
		}

		lines := 0
		exec := 0
		missing := []string{}

		for _, b := range p.Blocks {
			lines += b.NumStmt
			if b.Count > 0 {
				exec += b.NumStmt
			}

			if b.Count == 0 {
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

		printSummary(w,
			strings.TrimPrefix(p.FileName, base),
			lines, exec,
			strings.Join(missing, ","))

		totalLines += lines
		totalExec += exec
	}

	printLine(w)
	printSummary(w,
		"TOTAL",
		totalLines, totalExec,
		"")
}

func print(w *tabwriter.Writer, file, lines, exec, cover, missing string) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", file, lines, exec, cover, missing)
}

func printLine(w *tabwriter.Writer) {
	print(w, "", "", "", "", "")
}

func printSummary(w *tabwriter.Writer, name string, lines, exec int, missing string) {
	covered := float64(exec) / float64(lines)
	print(w,
		name,
		fmt.Sprintf("%d", lines),
		fmt.Sprintf("%d", exec),
		fmt.Sprintf("%0.1f%%", covered*100),
		missing)
}

func lcp(a, b string) string {
	min := a
	max := b

	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			return min[:i]
		}
	}

	return min
}
