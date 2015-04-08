package main

import (
	"flag"
	"fmt"

	"os"
	"os/exec"
)

const (
	coverOut = "cover.out"
)

var (
	filter = ".*"
)

func init() {
	flag.StringVar(&filter, "filter", filter, "filter which files to show")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	cleanup := false
	defer func() {
		if cleanup {
			os.Remove(coverOut)
		}
	}()

	file := coverOut

	if flag.NArg() > 0 {
		if flag.Arg(0) == "test" {
			args := flag.Args()
			args = append(args, "-coverprofile="+coverOut)
			cmd := exec.Command("go", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			cleanup = true
		} else {
			file = flag.Arg(0)
		}
	}

	dump(os.Stdout, os.Stderr, file, filter)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [option]... [<cover_file.out>|test]\n", os.Args[0])
    flag.PrintDefaults()
}
