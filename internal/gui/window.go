package gui

import "strings"

const (
	ld = "\u2514"
	rd = "\u2518"

	lu = "\u250C"
	ru = "\u2510"

	ho = "\u2500"
	ve = "\u2502"
)

func NewWindow(term *Terminal, r, c, rSize, cSize int) *Window {
	return &Window{
		term:    term,
		row:     r,
		col:     c,
		rowSize: rSize,
		colSize: cSize,
	}
}

type Window struct {
	term    *Terminal
	row     int
	col     int
	rowSize int
	colSize int
}

func (w *Window) Print() *Window {
	hor := strings.Repeat(ho, w.colSize)
	w.term.Goto(w.row, w.col)
	w.term.Print(lu, hor, ru)
	w.term.Goto(w.row+w.rowSize, w.col)
	w.term.Print(ld, hor, rd)

	for i := 1; i < w.rowSize; i++ {
		w.term.Goto(w.row+i, w.col).Print(ve)
		w.term.Goto(w.row+i, w.col+w.colSize+1).Print(ve)
	}
	return w
}

func (w *Window) Clear() *Window {
	hor := strings.Repeat(" ", w.colSize)
	for i := 1; i < w.rowSize; i++ {
		w.term.Goto(w.row+i, w.col+1).Print(hor)
	}
	return w
}

func (w *Window) GetRows(r, c int) (int, int) {
	return w.row + r, w.col + c
}
