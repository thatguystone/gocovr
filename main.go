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
	file   = coverOut
	filter = ".*"
)

func init() {
	flag.StringVar(&file, "file", file, "which coverage file to load")
	flag.StringVar(&filter, "filter", filter, "filter which files to show")
}

func main() {
	flag.Parse()

	cleanup := false
	defer func() {
		if cleanup {
			os.Remove(coverOut)
		}
	}()

	if flag.NArg() > 0 && flag.Arg(0) == "test" {
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
	}

	dump(os.Stdout, os.Stderr, file, filter)
}
