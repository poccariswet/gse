package main

import "fmt"

func main() {
	msg := "message"
	pos := 5
	msg = append(msg[:pos+1], msg[pos:]...)
	msg[pos] = "0"

	fmt.Println(msg)
}
