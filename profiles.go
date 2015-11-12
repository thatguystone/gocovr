package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/cover"
)

type profiles struct {
	sync.Mutex
	m map[string]*profile
}

type profile struct {
	sync.Mutex
	*cover.Profile
}

type profileSlice []*profile

func (ps *profiles) add(cp *cover.Profile) {
	ps.Lock()

	if ps.m == nil {
		ps.m = map[string]*profile{}
	}

	p := ps.m[cp.FileName]
	if p == nil {
		ps.m[cp.FileName] = &profile{
			Profile: cp,
		}
	}

	ps.Unlock()

	// First profile obviously doesn't need to be merged
	if p != nil {
		p.merge(cp)
	}
}

func (ps *profiles) getBase() string {
	base := ""

	for _, p := range ps.m {
		if base == "" {
			base = p.FileName
		} else {
			base = lcp(base, p.FileName)
		}
	}

	sep := fmt.Sprintf("%c", os.PathSeparator)
	if !strings.HasSuffix(base, sep) {
		base = filepath.Dir(base) + sep
	}

	return base
}

func (ps *profiles) sortedSlice() profileSlice {
	s := profileSlice{}

	for _, p := range ps.m {
		s = append(s, p)
	}

	sort.Sort(s)

	return s
}

// The coverage output for the same file should be identical across runs,
// assuming the file hasn't changed. If it's changed, then all bets are off.
func (p *profile) merge(cp *cover.Profile) {
	p.Lock()
	defer p.Unlock()

	if p.Mode != cp.Mode {
		panic(fmt.Errorf("profile mode mismatch: %s != %s",
			p.Mode,
			cp.Mode))
	}

	if len(p.Blocks) != len(cp.Blocks) {
		panic(fmt.Errorf("profile block len mismatch: %d != %d",
			len(p.Blocks),
			len(cp.Blocks)))
	}

	for i := 0; i < len(p.Blocks); i++ {
		pb := &p.Blocks[i]
		cpb := cp.Blocks[i]

		matches := pb.StartLine == cpb.StartLine &&
			pb.StartCol == cpb.StartCol &&
			pb.EndLine == cpb.EndLine &&
			pb.EndCol == cpb.EndCol &&
			pb.NumStmt == cpb.NumStmt
		if !matches {
			panic(fmt.Errorf("profile block range mismatch: %#v != %#v",
				pb,
				cpb))
		}

		pb.Count += cpb.Count
	}
}

func (ps profileSlice) Len() int           { return len(ps) }
func (ps profileSlice) Less(i, j int) bool { return ps[i].FileName < ps[j].FileName }
func (ps profileSlice) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
