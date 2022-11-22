package lib

import "fmt"

const ColorReset = "\033[0m"
const ColorBlue = "\033[34m"
const ColorRed = "\033[31m"
const ColorGreen = "\033[32m"
const ColorCyan = "\033[36m"

func Red(format string, a ...any) string {
	return fmt.Sprintf(ColorRed+format+ColorReset, a...)
}

func Green(format string, a ...any) string {
	return fmt.Sprintf(ColorGreen+format+ColorReset, a...)
}

func Blue(format string, a ...any) string {
	return fmt.Sprintf(ColorBlue+format+ColorReset, a...)
}

func Cyan(format string, a ...any) string {
	return fmt.Sprintf(ColorCyan+format+ColorReset, a...)
}
