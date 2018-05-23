package main

import (
	"fmt"
	"os"

	gc "github.com/rthornton128/goncurses"
)

const (
	ESC_KEY   gc.Key = 0x1B
	COLON_KEY gc.Key = 0x3A
)

var (
	quit = make(chan struct{})
)

func (v *View) ScrollWin() {
	if v.cursor.y > v.max_y {
		v.cursor.y = v.max_y
		if v.cursor.text_y < v.file.line_num && v.text_line != v.file.line_num {
			v.main_window.Scroll(1)
			v.colm_window.Scroll(1)
			v.colm_window.AttrOn(gc.A_BOLD)
			v.colm_window.MovePrintf(v.cursor.y, 0, "%3d ", v.cursor.text_y+1)
			v.colm_window.AttrOff(gc.A_BOLD)
			v.colm_window.Refresh()
			count := 0
			for {
				if v.max_x == count || v.file.bytes[v.text_pos] == byte(10) {
					v.main_window.MovePrint(v.cursor.y, count, string(v.file.bytes[v.text_pos]))
					v.text_pos++
					break
				}
				v.main_window.MovePrint(v.cursor.y, count, string(v.file.bytes[v.text_pos]))
				v.main_window.Refresh()
				v.text_pos++
				count++
			}
			v.main_window.Refresh()
		} else {
			v.cursor.text_y = v.file.line_num
		}
	}

	//TODO: scrol up できない
	if v.cursor.y < 0 {
		v.cursor.y = 0
		if v.cursor.text_y >= 0 {
			v.main_window.Scroll(-1)
			v.colm_window.Scroll(-1)
			v.colm_window.AttrOn(gc.A_BOLD)
			v.colm_window.MovePrintf(v.cursor.y, 0, "%3d ", v.cursor.text_y+1)
			v.colm_window.AttrOff(gc.A_BOLD)
			v.colm_window.Refresh()
			count := 0
			for {
				if v.max_x == count || v.file.bytes[v.text_pos] == byte(10) {
					v.main_window.MovePrint(v.cursor.y, count, string(v.file.bytes[v.text_pos]))
					v.text_pos++
					break
				}
				v.main_window.MovePrint(v.cursor.y, count, string(v.file.bytes[v.text_pos]))
				v.main_window.Refresh()
				v.text_pos++
				count++
			}
			v.main_window.Refresh()
		} else {
			v.cursor.text_y = 0
		}
	}

	//TODO: limit of x line

}

func (v *View) NormalCommand(ch gc.Key) {
	switch ch {
	case gc.KEY_RIGHT, 'l', gc.KEY_UP, 'k', gc.KEY_DOWN, 'j', '\n', gc.KEY_LEFT, 'h', gc.KEY_BACKSPACE:
		v.CursorMove(ch)

	case 'q':
		close(quit)
	case 'i':
		v.mode = Insert
	case 'v':
		v.mode = Visual
	}
}

func (v *View) Insert(ch string) {
	if ch == "\n" {

	}
}

//TODO
func (v *View) InsertCommand(ch gc.Key) {
	switch ch {
	case ESC_KEY:
		v.mode = Normal
	case gc.KEY_RIGHT, gc.KEY_UP, gc.KEY_DOWN, gc.KEY_LEFT:
		v.CursorMove(ch)
	default:
		v.Insert(string(ch))
	}
}

func (v *View) VisualCommand(ch gc.Key) {
	switch ch {
	case gc.KEY_RIGHT, 'l', gc.KEY_UP, 'k', gc.KEY_DOWN, 'j', '\n', gc.KEY_LEFT, 'h':
		v.CursorMove(ch)
	case ESC_KEY:
		v.mode = Normal
	}

}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, "Usage: command <filename>\n")
		os.Exit(1)
	}

	_, err := gc.Init()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	gc.StartColor() // start_color
	defer gc.End()  // endwin

	v, err := NewView()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err := v.file.Open(os.Args[1]); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err := v.Init(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	go func() {
		for {
			v.Mode()
			ch := v.main_window.GetChar()
			switch v.mode {
			case Normal:
				v.NormalCommand(ch)
			case Insert:
				v.InsertCommand(ch)
			case Visual:
				v.VisualCommand(ch)
			default:
				return
			}
		}
	}()

loop:
	for {
		select {
		case <-quit:
			break loop
		}
	}
}
