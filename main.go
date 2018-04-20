package main

import (
	"fmt"
	"log"

	gc "github.com/rthornton128/goncurses"
)

func main() {
	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	defer gc.End()

	stdscr.Refresh()
	for {
		ch := stdscr.GetChar()
		if ch == 'q' {
			break
		}
	}
	fmt.Println("good night")
}
