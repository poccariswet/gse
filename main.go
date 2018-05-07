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
	//	colm_window *gc.Window
	//	mode_window *gc.Window
}

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
	line, x := v.main_window.MaxYX() // ncurses_getmaxyx
	if line > len(contents) {
		line = len(contents)
	}
	v.cursor.max_y = line - 1
	v.cursor.max_x = x - 1

	for i := 0; i < line; i++ {
		v.main_window.Print(contents[i])
		v.main_window.Refresh()
	}
	v.main_window.Move(0, 0) // init locate of cursor
	v.main_window.Resize(line, x)
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
		if v.cursor.x < v.cursor.max_x {
			v.cursor.x++
		}
	case gc.KEY_UP, 'k':
		if v.cursor.y > 0 {
			v.cursor.y--
		}
	case gc.KEY_DOWN, 'j', '\n':
		if v.cursor.y < v.cursor.max_y {
			v.cursor.y++
		}
	}
	v.main_window.Move(v.cursor.y, v.cursor.x)
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
		//	case ColmWin:
		//		v.colm_window = win
		//	case ModeWin:
		//		v.mode_window = win
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
	if len_str < 3 {
		len_str = 3
	}
	y, x := stdscr.MaxYX()

	if err := v.MakeWindows(MainWin, y, x-len_str, 0, len_str); err != nil {
		return nil, err
	}

	//	if err := v.MakeWindows(ColmWin, y, len_str, 0, 0); err != nil {
	//		return nil, err
	//	}
	//
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
	if err := view.Init(f.contents); err != nil {
		log.Fatal(err)
	}

	quit := make(chan struct{})
	go func() {
		for {
			ch := view.main_window.GetChar()
			if ch == 'q' {
				close(quit)
			}
			switch view.mode {
			case Normal:
				if err := view.NormalCommand(ch); err != nil {
					log.Fatal(err)
				}
			case Insert:
			case Visual:
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
