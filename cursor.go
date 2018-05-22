package main

import gc "github.com/rthornton128/goncurses"

type Cursor struct {
	x      int
	y      int
	text_x int
	text_y int
}

func (v *View) CursorMove(ch gc.Key) {
	switch ch {
	case gc.KEY_LEFT, 'h', gc.KEY_BACKSPACE:
		if v.cursor.x > 0 {
			v.cursor.x--
			v.cursor.text_x--
		}
	case gc.KEY_RIGHT, 'l':
		if v.cursor.x < v.max_x {
			v.cursor.x++
			v.cursor.text_x++
		}
	case gc.KEY_UP, 'k':
		v.cursor.y--
		v.cursor.text_y--
	case gc.KEY_DOWN, 'j', '\n':
		v.cursor.y++
		v.cursor.text_y++
	}
	v.ScrollWin()
	v.main_window.Move(v.cursor.y, v.cursor.x)
}
