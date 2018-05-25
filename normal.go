package main

import gc "github.com/rthornton128/goncurses"

func (v *View) NormalCommand(ch gc.Key) {
	switch ch {
	case gc.KEY_RIGHT, 'l', gc.KEY_UP, 'k', gc.KEY_DOWN, 'j', '\n', gc.KEY_LEFT, 'h', DELETE_KEY:
		v.CursorMove(ch)

	case CTRS_KEY:
		v.Save()

	case 'q':
		close(quit)
	case 'i':
		v.mode = Insert
		v.state = false
	case 'v':
		v.mode = Visual
	}
}
