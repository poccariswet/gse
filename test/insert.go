package main

import (
	"fmt"
	"math"
	"os"
)

const (
	BUFSIZE = math.MaxInt32
)

func main() {

	file, err := os.Open("../test.txt")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	buf := make([]byte, BUFSIZE)
	n, err := file.Read(buf)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	buf = buf[:n]

	pos := 10
	buf = append(buf[:pos+1], buf[pos:]...)
	buf[pos] = 93 // 93 = ]

	fmt.Println(string(buf))
}
