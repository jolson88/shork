package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()
	fmt.Println("Press Ctrl+C to quit\r")

	var buf [1]byte
	for {
		n, err := os.Stdin.Read(buf[:])
		if err != nil || n == 0 {
			break
		}
		b := buf[0]
		if b == 0x03 { // Ctrl+C
			break
		}
		if b >= 32 && b <= 126 {
			fmt.Printf("0x%02X  %q\r\n", b, b)
		} else {
			fmt.Printf("0x%02X\r\n", b)
		}
	}
}
