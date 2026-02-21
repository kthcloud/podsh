package notice

import (
	"fmt"
	"strings"
)

type Level int

const (
	Info Level = iota
	Warning
	Error
	Success
)

func colorFor(l Level) string {
	switch l {
	case Info:
		return cyan
	case Warning:
		return yellow
	case Error:
		return red
	case Success:
		return green
	default:
		return blue
	}
}

func Show(level Level, title, message string) {
	width := termWidth()
	textWidth := width - 6
	color := colorFor(level)

	// top
	fmt.Println(color + "╷" + reset)

	// title
	fmt.Printf("%s│ %s%s%s\n", color, bold, title, reset)
	fmt.Printf("%s│%s\n", color, reset)

	// body
	for _, paragraph := range strings.Split(message, "\n") {
		lines := wrap(paragraph, textWidth)

		for _, l := range lines {
			fmt.Printf("%s│ %s\n", color, l)
		}
		fmt.Printf("%s│%s\n", color, reset)
	}

	// bottom
	fmt.Println(color + "╵" + reset)
}

func InfoBox(title, msg string) {
	Show(Info, title, msg)
}

func Warn(title, msg string) {
	Show(Warning, title, msg)
}

func Fail(title, msg string) {
	Show(Error, title, msg)
}

func SuccessBox(title, msg string) {
	Show(Success, title, msg)
}
