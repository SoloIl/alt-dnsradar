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

func colorizeTCPLabel(label string) string {
	switch label {
	case "BLOCKED":
		return colorize(label, ansiRed)
	case "-":
		return label
	default:
		return colorize(label, ansiGreen)
	}
}

func colorizeTLSStatus(status TLSStatus) string {
	switch status {
	case TLSStatusOK:
		return colorize(string(status), ansiGreen)
	case TLSStatusFail, TLSStatusTimeout:
		return colorize(string(status), ansiRed)
	default:
		return string(status)
	}
}
