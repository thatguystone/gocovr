package main

//gocovr:skip-file

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
	"sync"

	"os"
)

const (
	coverOut = "cover.out"
)

var (
	parallel = 1
	filter   = ".*"

	outCovRe = regexp.MustCompile(`\t?coverage: \d*\.\d*% of statements`)

	parallelCh chan struct{}
)

func init() {
	flag.IntVar(&parallel, "parallel", runtime.GOMAXPROCS(-1),
		"how many sub tests to run in parallel")
	flag.StringVar(&filter, "filter", filter,
		"filter which files to show")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if parallel < 1 {
		parallel = 1
	}

	parallelCh = make(chan struct{}, parallel)
	for i := 0; i < parallel; i++ {
		parallelCh <- struct{}{}
	}

	files := []string{}
	cleanup := false

	defer func() {
		if cleanup {
			for _, f := range files {
				os.Remove(f)
			}
		}
	}()

	var errs []error

	if flag.NArg() > 0 {
		if flag.Arg(0) == "test" {
			cleanup = true
			files, errs = runMultiple(flag.Args()[1:])
		} else {
			files = append(files, flag.Arg(0))
		}
	} else {
		files = append(files, coverOut)
	}

	if len(errs) == 0 {
		errs = dump(os.Stdout, files, filter)
	}

	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [option]... [<cover_file.out>|test]\n", os.Args[0])
	flag.PrintDefaults()
}

func runParallel() func() {
	<-parallelCh
	return func() {
		parallelCh <- struct{}{}
	}
}

func runCmd(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	return cmd.CombinedOutput()
}

func runMultiple(args []string) (coverfiles []string, errs []error) {
	cmds, coverfiles, errs := buildCmds(args)
	if len(errs) > 0 {
		return
	}

	type testStatus struct {
		pkg string
		out []byte
		err error
	}

	statusChs := make([]chan testStatus, len(cmds))

	for i, cmd := range cmds {
		ch := make(chan testStatus)
		statusChs[i] = ch

		go func(cmd []string, statusCh chan<- testStatus) {
			done := runParallel()
			out, err := runCmd(cmd...)
			done()

			statusCh <- testStatus{
				pkg: cmd[len(cmd)-1],
				out: out,
				err: err,
			}
		}(cmd, ch)
	}

	for _, ch := range statusChs {
		status := <-ch

		if status.err != nil && len(status.out) == 0 {
			errs = append(errs, fmt.Errorf("failed to run %s: %v", status.pkg, status.err))
		} else {
			out := outCovRe.ReplaceAll(status.out, []byte{})
			os.Stdout.Write(out)
		}
	}

	return
}

func buildCmds(args []string) (cmds [][]string, coverfiles []string, errs []error) {
	var mtx sync.Mutex

	addCmd := func(coverfile string, args []string) {
		mtx.Lock()
		coverfiles = append(coverfiles, coverfile)
		cmds = append(cmds, args)
		mtx.Unlock()
	}

	addErr := func(err error) {
		mtx.Lock()
		errs = append(errs, err)
		mtx.Unlock()
	}

	pkgSet := map[string]struct{}{}
	addPkg := func(pkg string) (ok bool) {
		mtx.Lock()
		_, has := pkgSet[pkg]
		pkgSet[pkg] = struct{}{}
		mtx.Unlock()

		ok = !has

		return
	}

	var flags []string
	pkgs := parsePkgs(args)

	if len(pkgs) == 0 {
		pkgs = append(pkgs, ".")
	} else {
		flags = args[:len(args)-len(pkgs)]
	}

	for _, pkg := range pkgs {
		stdout, err := runCmd("go", "list", pkg)
		if err != nil {
			addErr(fmt.Errorf("failed to list pkg %s: %v", pkg, err))
			return
		}

		sc := bufio.NewScanner(bytes.NewReader(stdout))
		for sc.Scan() {
			pkg := sc.Text()
			if !addPkg(pkg) {
				continue
			}

			f, err := ioutil.TempFile("", "gocovr-")
			if err != nil {
				addErr(fmt.Errorf("failed to create tmp cover file for %s: %v", pkg, err))
				return
			}

			cmd := []string{"go", "test", "-coverprofile=" + f.Name()}
			cmd = append(cmd, flags...)
			cmd = append(cmd, pkg)

			addCmd(f.Name(), cmd)
			f.Close()
		}
	}

	return
}

func parsePkgs(args []string) []string {
	set := flag.NewFlagSet("tmp parse", flag.ContinueOnError)

	// Ignore all test flags: just get position args
	set.Bool("a", false, "")
	set.Bool("race", false, "")
	set.Bool("short", false, "")
	set.Bool("v", false, "")
	set.Bool("x", false, "")
	set.String("covermode", "", "")
	set.String("coverprofile", "", "")
	set.String("cpu", "", "")
	set.String("parallel", "", "")
	set.String("run", "", "")
	set.String("timeout", "", "")

	set.Parse(args)

	return set.Args()
}
