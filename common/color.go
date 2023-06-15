// Package common is library function for whole system.
// # This manifest was generated by ymir. DO NOT EDIT.
package common

import "fmt"

const (
	ColorBlack = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite

	ColorBold     = 1
	ColorDarkGray = 90
)

// Colorize function for send (colored or common) message to output.
func Colorize(msg string, c int) string {
	// Send common or colored caption.
	//return fmt.Sprintf("\033[0;%d%s\033[0m", c, msg)
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, msg)
}

// ColorizeLevel function for send (colored or common) message to output.
func ColorizeLevel(level string) string {
	var (
		icon string
		c    int
	)
	// Switch color.
	switch level {
	case "success":
		icon = "[OK]"
		c = ColorGreen
	case "error":
		icon = "[ERROR]"
		c = ColorRed
	case "info":
		icon = "[INFO]"
		c = ColorBlue
	}
	return Colorize(icon, c)
}