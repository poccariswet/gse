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
	case gc.KEY_LEFT, 'h', DELETE_KEY:
		v.cursor.x--
		v.cursor.text_x--
	case gc.KEY_RIGHT, 'l':
		v.cursor.x++
		v.cursor.text_x++
	case gc.KEY_UP, 'k':
		v.cursor.y--
		v.cursor.text_y--
	case gc.KEY_DOWN, 'j', '\n':
		v.cursor.y++
		v.cursor.text_y++
	}
	v.ScrollWin(false)
	v.main_window.Move(v.cursor.y, v.cursor.x)
}

func (v *View) ScrollWin(scrol bool) {
	if v.cursor.y > v.max_y || scrol {
		if v.cursor.y > v.max_y {
			v.cursor.y = v.max_y
		}
		if v.cursor.text_y < v.file.GetLine() {
			v.main_window.Scroll(1)
			v.colm_window.Scroll(1)
			v.colm_window.AttrOn(gc.A_BOLD)
			v.colm_window.MovePrintf(v.cursor.y, 0, "%3d ", v.cursor.text_y+1)
			v.colm_window.AttrOff(gc.A_BOLD)
			v.main_window.Refresh()
			v.main_window.MovePrint(v.cursor.y, 0, v.file.buf[v.cursor.text_y])
			v.colm_window.Refresh()
		} else {
			v.cursor.text_y = v.file.GetLine()
		}
	}

	if v.cursor.y < 0 {
		v.cursor.y = 0
		if v.cursor.text_y >= 0 {
			v.main_window.Scroll(-1)
			v.colm_window.Scroll(-1)
			v.colm_window.AttrOn(gc.A_BOLD)
			v.colm_window.MovePrintf(v.cursor.y, 0, "%3d ", v.cursor.text_y+1)
			v.colm_window.AttrOff(gc.A_BOLD)
			v.main_window.Refresh()
			v.main_window.MovePrint(v.cursor.y, 0, v.file.buf[v.cursor.text_y])
			v.colm_window.Refresh()
		} else {
			v.cursor.text_y = 0
		}
	}

	if v.cursor.x < 0 {
		v.cursor.x = 0
	}

	if v.cursor.x > v.file.GetCol(v.cursor.y) {
		v.cursor.x = v.file.GetCol(v.cursor.y)
		if v.cursor.x < 0 {
			v.cursor.x = 0
		}
	}

}
