package main

import gc "github.com/rthornton128/goncurses"

type Mode int

const (
	Normal Mode = iota
	Insert
	Visual
)

func (m Mode) String() string {
	switch m {
	case Normal:
		return "Normal"
	case Insert:
		return "Insert"
	case Visual:
		return "Visual"
	default:
		return "non-match"
	}
}

func (v *View) Mode() {
	v.mode_window.AttrOn(gc.A_BOLD)
	v.mode_window.MovePrintf(0, 0, "%s", v.mode)
	v.mode_window.AttrOff(gc.A_BOLD)
	v.mode_window.MovePrintf(0, 6, ": %s\n", v.file.name)
	v.mode_window.Refresh()
	v.colm_window.Refresh()
	v.main_window.Refresh()
}
