package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/cover"
	"golang.org/x/tools/go/packages"
)

type profiles []*profile

func (s profiles) Len() int           { return len(s) }
func (s profiles) Less(i, j int) bool { return s[i].filename < s[j].filename }
func (s profiles) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s profiles) getBase() string {
	base := ""

	for _, p := range s {
		if base == "" {
			base = p.filename
		} else {
			base = lcp(base, p.filename)
		}
	}

	sep := fmt.Sprintf("%c", os.PathSeparator)
	if !strings.HasSuffix(base, sep) {
		base = filepath.Dir(base) + sep
	}

	return base
}

type profile struct {
	filename    string
	exec, total int
	missing     []string
}

type profilesMaker struct {
	covProfs    []*cover.Profile
	sourceFiles map[string]string

	mtx sync.Mutex
	s   profiles
}

func makeProfiles(file string) (profiles, error) {
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

	sort.Sort(pm.s)

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
		covProf := covProf

		g.Go(func() error {
			return pm.addProfile(covProf)
		})
	}

	return g.Wait()
}

func (pm *profilesMaker) addProfile(covProf *cover.Profile) error {
	absPath, ok := pm.sourceFiles[covProf.FileName]
	if !ok {
		return fmt.Errorf("could not locate source file for %s", covProf.FileName)
	}

	ignore, err := pm.ignoreFile(absPath)
	if ignore || err != nil {
		return err
	}

	p := profile{
		filename: covProf.FileName,
	}

	for _, b := range pm.coalesce(covProf.Blocks) {
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

func (*profilesMaker) coalesce(bs []cover.ProfileBlock) (res []cover.ProfileBlock) {
	for _, b := range bs {
		if len(res) > 0 {
			prev := &res[len(res)-1]

			// Two "misses" next to each other can always be joined
			if b.Count == 0 && prev.Count == 0 {
				prev.EndLine = b.EndLine
				prev.EndCol = b.EndCol
				prev.NumStmt += b.NumStmt
				prev.Count += b.Count
				continue
			}
		}

		res = append(res, b)
	}

	return
}

func (*profilesMaker) ignoreFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "//gocovr:skip-file" {
			return true, nil
		}
	}

	return false, s.Err()
}
