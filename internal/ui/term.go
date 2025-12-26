package ui

import (
	"fmt"
	"os"
)

func Clear() {
	fmt.Print("\033[H\033[2J\033[3J")
}

func WaitEnter() {
	fmt.Print("\nНажмите Enter, чтобы продолжить...")
	var b [1]byte
	_, _ = os.Stdin.Read(b[:])
}

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func Green(s string) string { return "\033[32m" + s + "\033[0m" }
func Red(s string) string   { return "\033[31m" + s + "\033[0m" }
func Cyan(s string) string  { return "\033[36m" + s + "\033[0m" }
func Bold(s string) string  { return "\033[1m" + s + "\033[0m" }


