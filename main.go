package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	gc "github.com/rthornton128/goncurses"
)

const (
	ESC_KEY    gc.Key = 0x1B
	COLON_KEY  gc.Key = 0x3A
	DELETE_KEY gc.Key = 0x7F
	CTRS_KEY   gc.Key = 0x13 // for save
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
}

var (
	quit = make(chan struct{})
)

// windowの設定、ファイルの表示をする
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
	if y > len(v.file.buf) {
		y = len(v.file.buf)
	} else {
		y -= 1
	}

	v.max_y = y - 1
	v.max_x = x - 1

	for i := 0; i < y; i++ {
		v.colm_window.AttrOn(gc.A_BOLD)
		v.colm_window.Printf("%3d ", i+1)
		v.colm_window.AttrOff(gc.A_BOLD)
		v.colm_window.Refresh()
		v.main_window.Print(v.file.buf[i])
		v.main_window.Refresh()
	}
	v.colm_window.Refresh()
	v.main_window.Move(0, 0) // init locate of cursor
	v.main_window.Refresh()

	return nil
}

func (v *View) refresh() {
	line := v.cursor.text_y
	for i := v.cursor.y; i < v.max_y; i++ {
		v.colm_window.AttrOn(gc.A_BOLD)
		v.colm_window.Printf("%3d ", i+1)
		v.colm_window.AttrOff(gc.A_BOLD)
		v.colm_window.Refresh()
		v.main_window.Print(v.file.buf[line])
		v.main_window.Refresh()
		line++
	}
}

func (v *View) Save() {

}

func (v *View) cmdinsert() {

}

func (v *View) CmdlineCommand(ch gc.Key) {
	switch ch {
	case DELETE_KEY:
		v.mode = Normal
		v.cmdinsert()
	case 'w':
		v.Save()
	}
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

func NewView(f *FileInfo) (*View, error) {
	v := &View{
		file:   f,
		cursor: Cursor{x: 0, y: 0, text_x: 0, text_y: 0},
		mode:   Normal,
	}

	stdscr := gc.StdScr()
	len_str := len(fmt.Sprintf("%d", v.file.GetLine()))
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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: command <filename>\n")
		os.Exit(1)
	}

	_, err := gc.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	gc.StartColor() // start_color
	defer gc.End()  // endwin

	f, err := OpenFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	view, err := NewView(f) //ここでサブウィンドウの作成
	if err != nil {
		log.Fatal(err)
	}
	if err := view.Init(); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			view.Mode()
			ch := view.main_window.GetChar()
			switch view.mode {
			case Normal:
				view.NormalCommand(ch)
			case Insert:
				view.InsertCommand(ch)
			case Visual:
				view.VisualCommand(ch)
			case Cmdline:
				view.CmdlineCommand(ch)
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
