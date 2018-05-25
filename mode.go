package main

import gc "github.com/rthornton128/goncurses"

type Mode int

const (
	Normal Mode = iota
	Insert
	Visual
	Cmdline
)

func (m Mode) String() string {
	switch m {
	case Normal:
		return "NORMAL"
	case Insert:
		return "INSERT"
	case Visual:
		return "VISUAL"
	case Cmdline:
		return "NORMAL"
	default:
		return "non-match"
	}
}

func (v *View) Mode() {
	v.mode_window.AttrOn(gc.A_BOLD)
	v.mode_window.MovePrintf(0, 0, "%s", v.mode)
	v.mode_window.AttrOff(gc.A_BOLD)
	if v.state {
		v.mode_window.MovePrintf(0, 6, ": %s cursor: x.%d y.%d, text_y.%d      Save!!\n", v.file.name, v.cursor.x, v.cursor.y, v.cursor.text_x, v.cursor.text_y)
	} else {
		v.mode_window.MovePrintf(0, 6, ": %s cursor: x.%d y.%d, text_y.%d\n", v.file.name, v.cursor.x, v.cursor.y, v.cursor.text_x, v.cursor.text_y)
	}
	v.mode_window.Refresh()
	v.colm_window.Refresh()
	v.main_window.Refresh()
}
