package main

import (
	"errors"
	"fmt"

	gc "github.com/rthornton128/goncurses"
)

type WindowMode int

const (
	MainWin WindowMode = iota
	ColmWin
	ModeWin
)

type View struct {
	cursor      Cursor
	mode        Mode
	file        *FileInfo
	main_window *gc.Window
	colm_window *gc.Window
	mode_window *gc.Window
	max_x       int
	max_y       int
	text_pos    int
	text_line   int
}

func NewView() (*View, error) {
	v := &View{
		file:      &FileInfo{},
		cursor:    Cursor{x: 0, y: 0, text_x: 0, text_y: 0},
		mode:      Normal,
		text_pos:  0,
		text_line: 0,
	}

	stdscr := gc.StdScr()
	len_str := len(fmt.Sprintf("%d", v.file.line_num))
	if len_str < 4 {
		len_str = 4
	}
	y, x := stdscr.MaxYX()

	if err := v.MakeWindows(MainWin, y-1, x-len_str, 0, len_str); err != nil {
		return nil, err
	}

	if err := v.MakeWindows(ColmWin, y-1, len_str, 0, 0); err != nil {
		return nil, err
	}

	if err := v.MakeWindows(ModeWin, 1, x, y-2, 0); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *View) MakeWindows(wm WindowMode, nline, ncolm, begin_y, begin_x int) error {
	win, err := gc.NewWindow(nline, ncolm, begin_y, begin_x)
	if err != nil {
		return err
	}
	switch wm {
	case MainWin:
		v.main_window = win
	case ColmWin:
		v.colm_window = win
	case ModeWin:
		v.mode_window = win
	default:
		return errors.New("invalid mode window")
	}

	return nil
}

func (v *View) Init() error {
	gc.Raw(true) // raw mode
	gc.Echo(false)
	if err := gc.HalfDelay(20); err != nil {
		return err
	}

	gc.MouseMask(gc.M_ALL, nil)
	v.main_window.Keypad(true)
	v.main_window.ScrollOk(true)
	v.colm_window.ScrollOk(true)
	gc.InitPair(1, gc.C_BLACK, gc.C_WHITE)
	v.mode_window.SetBackground(gc.ColorPair(1))
	y, x := v.main_window.MaxYX() // ncurses_getmaxyx
	if y > v.file.line_num {
		y = v.file.line_num + 1
	} else {
		y -= 1
	}

	v.max_y = y - 1
	v.max_x = x - 1

	v.Show(v.file.bytes)
	v.main_window.Move(0, 0) // init locate of cursor
	v.main_window.Refresh()

	return nil
}

func (v *View) ShowLine() {
	for i := 0; i < v.max_y; i++ {
		v.colm_window.AttrOn(gc.A_BOLD)
		v.colm_window.MovePrintf(i, 0, "%3d ", i+1)
		v.colm_window.AttrOff(gc.A_BOLD)
		v.colm_window.Refresh()
	}
}

func (v *View) Show(bytes []byte) {
	max_win := len(bytes)

	v.ShowLine()

	count := 0
	h := 0
	for i := 0; i < max_win; i++ {
		if v.max_x == count || bytes[i] == byte(10) {
			h++
			count = 0
			v.main_window.MovePrint(h, count, string(bytes[v.text_pos]))
			v.main_window.Refresh()
			v.text_pos++
			v.text_line++
			if v.max_y == v.text_line {
				break
			}
			continue
		}
		v.main_window.MovePrint(h, count, string(bytes[v.text_pos]))
		v.main_window.Refresh()
		count++
		v.text_pos++
	}

	v.max_y -= 1

	v.colm_window.Refresh()
	v.main_window.Refresh()
}
