package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	gc "github.com/rthornton128/goncurses"
)

type FileInfo struct {
	name     string
	namepath string
	file     *os.File
	buf      []string
}

type Cursor struct {
	x      int
	y      int
	text_x int
	text_y int
}

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

type WindowMode int

const (
	ESC_KEY   gc.Key = 0x1B
	COLON_KEY gc.Key = 0x3A
)

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

func OpenFile(filename string) (*FileInfo, error) {
	name, err := homedir.Expand(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", name)
	}

	scanner := bufio.NewScanner(file)
	var str []string
	for scanner.Scan() {
		str = append(str, scanner.Text()+"\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("scanner err:", err))
	}

	return &FileInfo{
		name:     file.Name(),
		namepath: name,
		file:     file,
		buf:      str,
	}, nil
}

func (f *FileInfo) GetLine() int {
	return len(f.buf)
}

func (f *FileInfo) GetName() string {
	return f.file.Name()
}

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

func (v *View) CursorMove(ch gc.Key) {
	switch ch {
	case gc.KEY_LEFT, 'h':
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
	v.ScrollWin(false)
	v.main_window.Move(v.cursor.y, v.cursor.x)
}

func (v *View) ScrollWin(scrol bool) {
	if v.cursor.y > v.max_y || scrol == true {
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

	//TODO: x軸のcursorの動きを制限

}

func (v *View) NormalCommand(ch gc.Key) {
	switch ch {
	case gc.KEY_RIGHT, 'l', gc.KEY_UP, 'k', gc.KEY_DOWN, 'j', '\n', gc.KEY_LEFT, 'h':
		v.CursorMove(ch)

	case 'q':
		close(quit)
	case 'i':
		v.mode = Insert
	case 'v':
		v.mode = Visual
	}
}

func (v *View) refresh() {
	y, _ := v.main_window.MaxYX() // ncurses_getmaxyx
	if y > len(v.file.buf) {
		y = len(v.file.buf)
	} else {
		y -= 1
	}

	for i := 0; i < y; i++ {
		v.colm_window.AttrOn(gc.A_BOLD)
		v.colm_window.Printf("%3d ", i+1)
		v.colm_window.AttrOff(gc.A_BOLD)
		v.colm_window.Refresh()
		v.main_window.Print(v.file.buf[i])
		v.main_window.Refresh()
	}
}

func (v *View) insertNewLine() {
	y, _ := v.main_window.MaxYX() // ncurses_getmaxyx
	if y > len(v.file.buf) {
		y = len(v.file.buf)
	} else {
		y -= 1
	}

	str_copy := []byte(v.file.buf[v.cursor.text_y])
	oline := str_copy[0:v.cursor.text_x]
	nline := str_copy[v.cursor.text_x:]

	v.file.buf = append(v.file.buf[:v.cursor.text_y+1], v.file.buf[v.cursor.text_y:]...)
	v.file.buf[v.cursor.text_y] = string(oline)
	v.file.buf[v.cursor.text_y+1] = string(nline)
	v.cursor.text_y++
	v.cursor.y++
	v.ScrollWin(true)
	v.main_window.Move(v.cursor.y, v.cursor.x)
	v.max_y = y - 1
}

func (v *View) insert(ch gc.Key) {

}

func (v *View) Insert(ch gc.Key) {
	if ch == '\n' {
		v.insertNewLine()
	} else {
		v.insert(ch)
	}
}

//TODO: 文字の入力
func (v *View) InsertCommand(ch gc.Key) {
	switch ch {
	case ESC_KEY:
		v.mode = Normal
	case gc.KEY_RIGHT, gc.KEY_UP, gc.KEY_DOWN, gc.KEY_LEFT:
		v.CursorMove(ch)
	default:
		v.Insert(ch)
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

func (v *View) Mode() {
	v.mode_window.AttrOn(gc.A_BOLD)
	v.mode_window.MovePrintf(0, 0, "%s", v.mode)
	v.mode_window.AttrOff(gc.A_BOLD)
	v.mode_window.MovePrintf(0, 6, ": %s\n", v.file.name)
	v.mode_window.Refresh()
	v.colm_window.Refresh()
	v.main_window.Refresh()
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
