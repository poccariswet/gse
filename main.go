package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	gc "github.com/rthornton128/goncurses"
)

type FileConfig struct {
	file     *os.File
	contents []string
}

type Cursor struct {
	x int
	y int
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

type View struct {
	cursor Cursor
	mode   Mode
	window *gc.Window
}

func OpenFile(filename string) (*FileConfig, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
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

	return &FileConfig{
		file:     file,
		contents: str,
	}, nil
}

func (v *View) Init(contents []string) error {
	gc.Raw(true) // raw mode
	gc.Echo(false)
	if err := gc.HalfDelay(20); err != nil {
		return err
	}
	gc.MouseMask(gc.M_ALL, nil)
	v.window.Keypad(true)
	v.window.ScrollOk(true)
	line, _ := v.window.MaxYX() // ncurses_getmaxyx
	if line > len(contents) {
		line = len(contents)
	}

	for i := 0; i < line; i++ {
		v.window.Print(contents[i])
		//		v.window.Refresh()
	}
	v.window.Move(0, 0) // init locate of cursor
	v.window.Refresh()

	return nil
}

func (v *View) NormalCommand(ch gc.Key) error {
	switch ch {
	case gc.KEY_LEFT, 'h':
		v.cursor.x--
	case gc.KEY_RIGHT, 'l':
		v.cursor.x++
	case gc.KEY_UP, 'k':
		v.cursor.y--
	case gc.KEY_DOWN, 'j':
		v.cursor.y++
	}
	v.window.Move(v.cursor.y, v.cursor.x)
	return nil
}

func NewView(w *gc.Window) *View {
	return &View{
		cursor: Cursor{x: 0, y: 0},
		mode:   Normal,
		window: w,
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: command <filename>\n")
		os.Exit(1)
	}

	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	gc.StartColor() // start_color
	defer gc.End()  // endwin

	fc, err := OpenFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	view := NewView(stdscr)
	if err := view.Init(fc.contents); err != nil {
		log.Fatal(err)
	}

	for {
		ch := view.window.GetChar()
		if ch == 'q' {
			break
		}
		switch view.mode {
		case Normal:
			if err := view.NormalCommand(ch); err != nil {
				log.Fatal(err)
			}
		default:
			return
		}
	}
}
