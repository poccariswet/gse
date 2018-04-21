package main

import (
	"fmt"
)

type Mode int

const (
	Nomal Mode = iota
	Insert
)

func (m Mode) String() string {
	switch m {
	case Nomal:
		return "Nomal"
	case Insert:
		return "Insert"
	default:
		return "non-match"
	}
}

func main() {
	fmt.Println(Nomal)
}
