package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

type FileInfo struct {
	name     string
	namepath string
	file     *os.File
	buf      []string
}

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
		text := ""
		for _, v := range []byte(scanner.Text()) {
			if v == byte('\t') {
				text += "  "
				continue
			}
			text += string(v)
		}
		str = append(str, text+"\n")
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

func (f *FileInfo) GetCol(y int) int {
	if len(f.buf[y]) == 1 {
		return 0
	} else if len(f.buf[y]) == 0 {
		return 0
	}
	return len(f.buf[y]) - 1
}

func (f *FileInfo) GetName() string {
	return f.file.Name()
}
