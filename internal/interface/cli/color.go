package cli

import (
	"fmt"
	"os"
)

type Color string

const (
	ColorReset  Color = "[0m"
	ColorRed    Color = "[31m"
	ColorGreen  Color = "[32m"
	ColorYellow Color = "[33m"
	ColorBlue   Color = "[34m"
	ColorPurple Color = "[35m"
	ColorCyan   Color = "[36m"
	ColorWhite  Color = "[37m"
	ColorBold   Color = "[1m"
)

func Colorize(text string, color Color) string {
	return fmt.Sprintf("%s%s%s", string(color), text, string(ColorReset))
}

func PrintSuccess(text string) {
	fmt.Println(Colorize(text, ColorGreen))
}

func PrintError(text string) {
	fmt.Fprintln(os.Stderr, Colorize(text, ColorRed))
}

func PrintInfo(text string) {
	fmt.Println(Colorize(text, ColorBlue))
}

func PrintWarning(text string) {
	fmt.Println(Colorize(text, ColorYellow))
}

func PrintHeader(text string) {
	fmt.Println(Colorize(text, ColorBold))
}
