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
	namepath string
	file     *os.File
	contents []string
}

type Cursor struct {
	x     int
	y     int
	max_x int
	max_y int
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
	MainWin WindowMode = iota
	ColmWin
	ModeWin
)

type View struct {
	cursor      Cursor
	mode        Mode
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

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var str []string
	for scanner.Scan() {
		str = append(str, scanner.Text()+"\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("scanner err:", err))
	}

	return &FileInfo{
		namepath: name,
		file:     file,
		contents: str,
	}, nil
}

func (f *FileInfo) GetLine() int {
	return len(f.contents)
}

func (f *FileInfo) GetName() string {
	return f.file.Name()
}

// windowの設定、ファイルの表示をする
func (v *View) Init(contents []string) error {
	gc.Raw(true) // raw mode
	gc.Echo(false)
	if err := gc.HalfDelay(20); err != nil {
		return err
	}

	gc.MouseMask(gc.M_ALL, nil)
	v.main_window.Keypad(true)
	v.main_window.ScrollOk(true)
	v.colm_window.ScrollOk(true)
	y, x := v.main_window.MaxYX() // ncurses_getmaxyx
	if y > len(contents) {
		y = len(contents)
	} else {
		y -= 1
	}

	v.max_y = y - 1
	v.max_x = x - 1

	for i := 0; i < y; i++ {
		v.colm_window.Printf("%3d ", i+1)
		v.colm_window.Refresh()
		v.main_window.Print(contents[i])
		v.main_window.Refresh()
	}
	v.mode_window.Printf("%s", v.mode)
	v.mode_window.Refresh()
	v.colm_window.Refresh()
	v.main_window.Move(0, 0) // init locate of cursor
	v.main_window.Resize(y, x)
	v.main_window.Refresh()

	return nil
}

// Normal mode時のキー操作
func (v *View) NormalCommand(ch gc.Key) error {
	switch ch {
	case gc.KEY_LEFT, 'h':
		if v.cursor.x > 0 {
			v.cursor.x--
		}
	case gc.KEY_RIGHT, 'l':
		if v.cursor.x < v.max_x {
			v.cursor.x++
		}
	case gc.KEY_UP, 'k':
		if v.cursor.y > 0 {
			v.cursor.y--
		}
	case gc.KEY_DOWN, 'j', '\n':
		if v.cursor.y < v.max_y {
			v.cursor.y++
		}
	case 'q':
		close(quit)

	case 'i':
		v.mode = Insert
		v.mode_window.Printf("%s", v.mode)
		v.mode_window.Refresh()
	case 'v':
		v.mode = Visual
	}
	v.main_window.Move(v.cursor.y, v.cursor.x)
	return nil
}

func (v *View) InsertCommand(ch gc.Key) error {
	switch ch {
	case 'q':
		v.mode = Normal
		v.mode_window.Printf("%s", v.mode)
		v.mode_window.Refresh()
	case gc.KEY_LEFT:
		if v.cursor.x > 0 {
			v.cursor.x--
		}
	case gc.KEY_RIGHT:
		if v.cursor.x < v.max_x {
			v.cursor.x++
		}
	case gc.KEY_UP:
		if v.cursor.y > 0 {
			v.cursor.y--
		}
	case gc.KEY_DOWN:
		if v.cursor.y < v.max_y {
			v.cursor.y++
		}
	}
	v.main_window.Move(v.cursor.y, v.cursor.x)
	return nil
}

func (v *View) VisualCommand(ch gc.Key) error {

	return nil
}

// TODO サブウィンドウの追加
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
		cursor: Cursor{x: 0, y: 0},
		mode:   Normal,
	}

	stdscr := gc.StdScr()
	len_str := len(fmt.Sprintf("%d", f.GetLine()))
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

	if err := v.MakeWindows(ModeWin, 1, x, y-1, 0); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *View) Mode() {
	v.mode_window.MovePrint(v.max_y, 0, v.mode)
	v.mode_window.Refresh()
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
	if err := view.Init(f.contents); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			view.Mode()
			ch := view.main_window.GetChar()
			switch view.mode {
			case Normal:
				if err := view.NormalCommand(ch); err != nil {
					log.Fatal(err)
				}
			case Insert:
				if err := view.InsertCommand(ch); err != nil {
					log.Fatal(err)
				}
			case Visual:
				if err := view.VisualCommand(ch); err != nil {
					log.Fatal(err)
				}
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
