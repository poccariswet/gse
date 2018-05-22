package main

import (
	"fmt"
	"math"
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	BUFSIZE = math.MaxInt32
)

type FileInfo struct {
	name     string
	namepath string
	file     *os.File
	bytes    []byte
	size     int
	line_num int
}

func (f *FileInfo) Read() error {
	buf := make([]byte, BUFSIZE)
	n, err := f.file.Read(buf)
	if err != nil {
		return err
	}

	f.size = n
	f.bytes = buf[:n]

	return nil
}

func (f *FileInfo) Open(filename string) error {
	name, err := homedir.Expand(filename)
	if err != nil {
		return err
	}
	f.namepath = name

	fp, err := os.Open(name)
	if err != nil {
		return err
	}
	defer fp.Close()
	f.name = fp.Name()
	f.file = fp

	info, err := os.Stat(name)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", name)
	}

	if err := f.Read(); err != nil {
		return err
	}
	f.SetLine()

	return nil
}

func (f *FileInfo) SetLine() {
	line := 0

	for _, v := range f.bytes {
		if string(v) == "\n" {
			line++
		}
	}

	f.line_num = line
}

func (f *FileInfo) GetLine() int {
	return f.line_num
}

func (f *FileInfo) GetName() string {
	return f.name
}
