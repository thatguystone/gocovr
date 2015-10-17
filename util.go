package main

import "sync"

type parallelize struct {
	sync.Mutex
	sync.WaitGroup
	errs []error
}

func (pll *parallelize) do(ss []string, fn func(string) error) {
	for _, s := range ss {
		func(s string) {
			pll.Add(1)
			go func() {
				defer pll.Done()
				pll.addError(fn(s))
			}()
		}(s)
	}
}

func (pll *parallelize) addError(err error) {
	if err != nil {
		pll.Lock()
		pll.errs = append(pll.errs, err)
		pll.Unlock()
	}
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
