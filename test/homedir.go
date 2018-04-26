package main

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Printf("Usage: command <filename>\n")
		os.Exit(1)
	}

	path, err := homedir.Expand(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(path)

	path, err = homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(path)
}
