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

func Openfile(filename string) (*FileConfig, error) {
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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: command <filename>\n")
		os.Exit(1)
	}

	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	defer gc.End() //endwin

	fc, err := Openfile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range fc.contents {
		fmt.Print(v)
	}

	for {
		ch := stdscr.GetChar()
		if ch == 'q' {
			break
		}
	}
}
