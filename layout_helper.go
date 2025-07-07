package main

import (
	. "modernc.org/tk9.0"
)

type GridLayout struct {
	Rows    int
	Padding Opts
}
type Conf = func(col int) ([]Opt, int)

func ColSpan(n int, opts ...Opt) Conf {
	return func(i int) ([]Opt, int) { return append(opts, Columnspan(n)), n }
}
func (s *GridLayout) Row(val ...any) {
	r := Row(s.Rows)
	i := 0
	xi := 0
	var w Widget
	var o = append([]Opt{}, s.Padding...)
	for _, v := range val {
		switch x := v.(type) {
		case Widget:
			if w != nil {
				Grid(w, append(o, r, Column(i))...)
				i += xi + 1
			}
			w = x
			o = o[:0]
			o = append(o, s.Padding...)
		case Opt:
			o = append(o, x)
		case Conf:
			ox, d := x(i)
			xi += d
			o = append(o, ox...)
		}
	}
	Grid(w, append(o, r, Column(i))...)
	s.Rows++
}
func (s *GridLayout) ResetRow() {
	s.Rows = 0
}
