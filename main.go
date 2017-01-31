package main

//gocovr:skip-file

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"os"
)

const (
	coverOut = "cover.out"
)

var (
	parallel = 1
	include  = ".*"
	exclude  = "^$"

	cwd = ""

	outCovRe  = regexp.MustCompile(`\t?coverage: \d*\.\d*% of statements`)
	warningRe = regexp.MustCompile(`warning: no packages being tested depend on .*\n`)
)

func init() {
	flag.IntVar(&parallel, "parallel", runtime.GOMAXPROCS(-1),
		"how many sub tests to run in parallel")
	flag.StringVar(&include, "include", include,
		"which files to include")
	flag.StringVar(&exclude, "exclude", exclude,
		"which files to exclude")

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if parallel < 1 {
		parallel = 1
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
		fmt.Println()
		errs = dump(os.Stdout, files, include, exclude)
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

	parallelCh := make(chan struct{}, parallel)
	for i := 0; i < parallel; i++ {
		parallelCh <- struct{}{}
	}

	statusChs := make([]chan testStatus, len(cmds))

	for i, cmd := range cmds {
		ch := make(chan testStatus)
		statusChs[i] = ch

		go func(cmd []string, statusCh chan<- testStatus) {
			<-parallelCh
			out, err := runCmd(cmd...)
			parallelCh <- struct{}{}

			out = outCovRe.ReplaceAll(out, []byte{})
			out = warningRe.ReplaceAll(out, []byte{})

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
			os.Stdout.Write(status.out)
		}
	}

	return
}

func buildCmds(args []string) (cmds [][]string, coverfiles []string, errs []error) {
	pll := parallelize{}

	addCmd := func(coverfile string, args []string) {
		pll.Lock()
		coverfiles = append(coverfiles, coverfile)
		cmds = append(cmds, args)
		pll.Unlock()
	}

	pkgSet := map[string]struct{}{}
	addPkg := func(pkg string) (ok bool) {
		pll.Lock()
		_, has := pkgSet[pkg]
		pkgSet[pkg] = struct{}{}
		pll.Unlock()

		ok = !has

		return
	}

	var flags []string
	pkgs := parsePkgs(args)

	if len(pkgs) == 0 {
		pkgs = append(pkgs, ".")
		flags = args
	} else {
		flags = args[:len(args)-len(pkgs)]
	}

	pll.do(pkgs, func(pkg string) error {
		stdout, err := runCmd("go", "list", pkg)
		if err != nil {
			return fmt.Errorf("failed to list pkg %s: %v", pkg, err)
		}

		sc := bufio.NewScanner(bytes.NewReader(stdout))
		for sc.Scan() {
			pkg := sc.Text()
			if !addPkg(pkg) {
				continue
			}

			// If working in some path outside of GOPATH, the pkg path needs
			// to be made rel to the current path so that it still works.
			if strings.HasPrefix(pkg, "_/") {
				pkg = pkg[1:]
				relPkg, err := filepath.Rel(cwd, pkg)
				if err == nil {
					pkg = "./" + relPkg
				}
			}

			f, err := ioutil.TempFile("", "gocovr-")
			if err != nil {
				return fmt.Errorf("failed to create tmp cover file for %s: %v",
					pkg,
					err)
			}

			cmd := []string{"go", "test", "-coverprofile=" + f.Name()}
			cmd = append(cmd, flags...)
			cmd = append(cmd, pkg)

			addCmd(f.Name(), cmd)
			f.Close()
		}

		return nil
	})

	pll.Wait()
	errs = pll.errs

	return
}

func parsePkgs(args []string) []string {
	set := flag.NewFlagSet("tmp parse", flag.ContinueOnError)

	set.Usage = func() {
		cmd := exec.Command("go", "test", "-h")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		os.Exit(2)
	}

	// Ignore all test flags: just get position args
	set.Bool("a", false, "")
	set.Bool("i", false, "")
	set.Bool("race", false, "")
	set.Bool("short", false, "")
	set.Bool("v", false, "")
	set.Bool("x", false, "")
	set.String("covermode", "", "")
	set.String("coverpkg", "", "")
	set.String("coverprofile", "", "")
	set.String("cpu", "", "")
	set.String("parallel", "", "")
	set.String("run", "", "")
	set.String("timeout", "", "")

	set.Parse(args)

	return set.Args()
}
