package main

//gocovr:skip-file

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	status := run()
	os.Exit(status)
}

func run() int {
	showCovered := flag.Bool("showCovered", false, "show files with 100% coverage")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	coverFile := flag.Arg(0)

	if coverFile == "test" {
		f, err := os.CreateTemp("", "gocovr-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		f.Close()

		coverFile = f.Name()
		defer os.Remove(coverFile)

		args := []string{"test"}
		args = append(args, flag.Args()[1:]...)
		args = append(args, "-coverprofile="+coverFile)

		cmd := exec.Command("go", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			// No need to print error if the command exited; go will have
			// printed enough error info.
			if _, ok := err.(*exec.ExitError); !ok {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			return 1
		}
	}

	err := dump(os.Stdout, coverFile, *showCovered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [args] <subcommand>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "arguments:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "subcommands:\n")
	fmt.Fprintf(os.Stderr, "  test [go test args]  run go test\n")
	fmt.Fprintf(os.Stderr, "  <cover.out>          dump coverage file\n")
	os.Exit(2)
}
