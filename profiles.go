package main

import (
	"bufio"
	"bytes"
	"cmp"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/cover"
	"golang.org/x/tools/go/packages"
)

type profile struct {
	filename    string
	exec, total int
	missing     []string
}

type profilesMaker struct {
	covProfs    []*cover.Profile
	sourceFiles map[string]string

	mtx sync.Mutex
	s   []*profile
}

func makeProfiles(file string) ([]*profile, error) {
	covProfs, err := cover.ParseProfiles(file)
	if err != nil {
		return nil, fmt.Errorf("invalid coverage profile: %v", err)
	}

	pm := profilesMaker{
		covProfs: covProfs,
	}

	err = pm.loadPackageFiles()
	if err != nil {
		return nil, err
	}

	err = pm.addAllProfiles()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(pm.s, func(a, b *profile) int {
		return cmp.Compare(a.filename, b.filename)
	})

	return pm.s, nil
}

func (pm *profilesMaker) loadPackageFiles() error {
	var pkgs []string
	seen := make(map[string]struct{})

	for _, covProf := range pm.covProfs {
		pkg := filepath.Dir(covProf.FileName)

		if _, ok := seen[pkg]; !ok {
			seen[pkg] = struct{}{}
			pkgs = append(pkgs, pkg)
		}
	}

	res, err := packages.Load(nil, pkgs...)
	if err != nil {
		return fmt.Errorf("failed to locate packages: %v", err)
	}

	pm.sourceFiles = make(map[string]string)
	for _, pkg := range res {
		for _, f := range pkg.GoFiles {
			pkgFile := filepath.Join(pkg.PkgPath, filepath.Base(f))
			pm.sourceFiles[pkgFile] = f
		}
	}

	return nil
}

func (pm *profilesMaker) addAllProfiles() error {
	var g errgroup.Group

	for _, covProf := range pm.covProfs {
		g.Go(func() error {
			return pm.addProfile(covProf)
		})
	}

	return g.Wait()
}

func (pm *profilesMaker) addProfile(covProf *cover.Profile) error {
	srcPath, ok := pm.sourceFiles[covProf.FileName]
	if !ok {
		return fmt.Errorf("could not locate source file for %s", covProf.FileName)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	defer src.Close()

	ignore, err := ignoreFile(src)
	if ignore || err != nil {
		return err
	}

	// ignoreFile might read the whole file
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	p := profile{
		filename: covProf.FileName,
	}

	for b, err := range iterBlocks(covProf.Blocks, src) {
		if err != nil {
			return err
		}

		p.total += b.NumStmt
		if b.Count > 0 {
			p.exec += b.NumStmt
		}

		if b.Count == 0 && b.NumStmt > 0 {
			if b.StartLine == b.EndLine {
				p.missing = append(p.missing,
					fmt.Sprintf("%d", b.StartLine))
			} else {
				p.missing = append(p.missing,
					fmt.Sprintf("%d-%d", b.StartLine, b.EndLine))
			}
		}
	}

	pm.mtx.Lock()
	pm.s = append(pm.s, &p)
	pm.mtx.Unlock()

	return nil
}

// https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source/
var genRe = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func ignoreFile(src io.Reader) (bool, error) {
	sc := bufio.NewScanner(src)
	for sc.Scan() {
		l := sc.Text()
		done := strings.HasPrefix(l, "import ") ||
			strings.HasPrefix(l, "type ") ||
			strings.HasPrefix(l, "const ") ||
			strings.HasPrefix(l, "var ") ||
			strings.HasPrefix(l, "func ")
		if done {
			break
		}

		if genRe.MatchString(l) {
			return true, nil
		}

		if l == "//gocovr:skip-file" {
			return true, nil
		}
	}

	return false, sc.Err()
}

func iterBlocks(
	bs []cover.ProfileBlock,
	src *os.File,
) iter.Seq2[cover.ProfileBlock, error] {
	return func(yield func(cover.ProfileBlock, error) bool) {
		prev := bs[0]
		it := fileIter{sc: bufio.NewScanner(src)}
		for _, b := range bs[1:] {
			if b.Count == 0 {
				ignore, err := it.ignoreBlock(b)
				if err != nil {
					yield(b, err)
					return
				}

				if ignore {
					b.Count = 1
				}
			}

			// Two "misses" next to each other can always be joined
			if b.Count == 0 && prev.Count == 0 {
				prev.EndLine = b.EndLine
				prev.EndCol = b.EndCol
				prev.NumStmt += b.NumStmt
				prev.Count += b.Count
				continue
			}

			if !yield(prev, nil) {
				return
			}

			prev = b
		}

		yield(prev, nil)
	}
}

type fileIter struct {
	sc     *bufio.Scanner
	buf    bytes.Buffer
	lineno int
}

func (it *fileIter) scan() bool {
	ok := it.sc.Scan()
	if ok {
		it.lineno++
	}
	return ok
}

var unreachables = [][]byte{
	[]byte("assert.unreachable("),
	[]byte(`panic("unreachable`),
	[]byte("panic(`unreachable"),
	[]byte(`panic(fmt.errorf("unreachable`),
	[]byte("panic(fmt.errorf(`unreachable"),
	[]byte(`panic(fmt.sprintf("unreachable`),
	[]byte("panic(fmt.sprintf(`unreachable"),
}

func (it *fileIter) ignoreBlock(b cover.ProfileBlock) (bool, error) {
	for it.lineno < b.StartLine {
		if !it.scan() {
			return false, it.sc.Err()
		}
	}

	it.buf.Reset()
	add := func(b []byte) {
		b = bytes.TrimSpace(b)
		if bytes.HasPrefix(b, []byte("//")) {
			return
		}

		b = bytes.ToLower(b)
		it.buf.Write(b)
	}

	if b.StartLine == b.EndLine {
		add(it.sc.Bytes()[b.StartCol-1 : b.EndCol-1])
	} else {
		add(it.sc.Bytes()[b.StartCol-1:])

		for {
			if !it.scan() {
				return false, it.sc.Err()
			}

			if it.lineno == b.EndLine {
				add(it.sc.Bytes()[:b.EndCol-1])
				break
			}

			add(it.sc.Bytes())
		}
	}

	for _, s := range unreachables {
		if bytes.Contains(it.buf.Bytes(), s) {
			return true, nil
		}
	}

	return false, nil
}
