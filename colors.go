package main

import (
	"fmt"
	"os"
)

const (
	ansiReset = "\033[0m"
	ansiRed   = "\033[31m"
	ansiGreen = "\033[32m"
)

func colorsEnabled() bool {
	if flagSettings.NoColor != nil && *flagSettings.NoColor {
		return false
	}

	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func colorize(s string, color string) string {
	if !colorsEnabled() {
		return s
	}

	return fmt.Sprintf("%s%s%s", color, s, ansiReset)
}

func padAndColorize(s string, width int, color string) string {
	padded := fmt.Sprintf("%-*s", width, s)
	if color == "" {
		return padded
	}

	return colorize(padded, color)
}

func colorizeTCPLabel(label string) string {
	switch label {
	case "BLOCKED":
		return padAndColorize(label, 9, ansiRed)
	case "-":
		return padAndColorize(label, 9, "")
	default:
		return padAndColorize(label, 9, ansiGreen)
	}
}

func colorizeTLSStatus(status TLSStatus) string {
	switch status {
	case TLSStatusOK:
		return padAndColorize(string(status), 8, ansiGreen)
	case TLSStatusFail, TLSStatusTimeout:
		return padAndColorize(string(status), 8, ansiRed)
	default:
		return padAndColorize(string(status), 8, "")
	}
}
