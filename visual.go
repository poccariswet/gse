package main

import gc "github.com/rthornton128/goncurses"

func (v *View) VisualCommand(ch gc.Key) {
	switch ch {
	case gc.KEY_RIGHT, 'l', gc.KEY_UP, 'k', gc.KEY_DOWN, 'j', '\n', gc.KEY_LEFT, 'h':
		v.CursorMove(ch)
	case ESC_KEY:
		v.mode = Normal
	}

}
