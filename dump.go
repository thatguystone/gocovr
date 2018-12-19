package main

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func dump(out io.Writer, file string, showCovered bool) error {
	profs, err := makeProfiles(file)
	if err != nil {
		return err
	}

	if len(profs) == 0 {
		fmt.Fprintln(out, "No files covered.")
		return nil
	}

	base := profs.getBase()
	fmt.Fprintf(out, "\nBase: %s\n\n", base)

	w := newWriter(out)
	defer w.Flush()

	w.print("File", "Lines", "Exec", "Cover", "Missing")
	w.blank()

	totalLines := 0
	totalExec := 0

	for _, prof := range profs {
		if len(prof.missing) > 0 || showCovered {
			w.summary(
				strings.TrimPrefix(prof.filename, base),
				prof.total, prof.exec,
				strings.Join(prof.missing, ","))
		}

		totalLines += prof.total
		totalExec += prof.exec
	}

	w.blank()
	w.summary(
		"TOTAL",
		totalLines, totalExec,
		"")

	return nil
}

type writer struct {
	*tabwriter.Writer
}

func newWriter(w io.Writer) writer {
	return writer{tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)}
}

func (w writer) blank() {
	w.print("", "", "", "", "")
}

func (w writer) print(file, total, exec, cover, missing string) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", file, total, exec, cover, missing)
}

func (w writer) summary(name string, total, exec int, missing string) {
	covered := float64(exec) / float64(total)
	if exec == 0 && total == 0 {
		covered = 1
	}

	w.print(
		name,
		fmt.Sprintf("%d", total),
		fmt.Sprintf("%d", exec),
		fmt.Sprintf("%0.1f%%", covered*100),
		missing)
}
