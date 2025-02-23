package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
	rgbBlack  = "#000000"
	rgbGreen  = "#5CCD5C"
	rgbRed    = "#B22222"
	rgbWhite  = "#FFFFFF"
)

func PrettyError(err error) {
	fmt.Fprintf(os.Stderr, "%s❌ %s%s\n", ansiRed, err, ansiReset)
}

func PrettyPercentageTest(value, min int, msg string) {
	var color lipgloss.AdaptiveColor
	switch {
	case min == 0:
		color = lipgloss.AdaptiveColor{Light: rgbBlack, Dark: rgbWhite}
	case min <= value:
		color = lipgloss.AdaptiveColor{Light: rgbGreen, Dark: rgbGreen}
	case value < min:
		color = lipgloss.AdaptiveColor{Light: rgbRed, Dark: rgbRed}
	default:
		color = lipgloss.AdaptiveColor{Light: rgbBlack, Dark: rgbWhite}
	}

	totalBars := 70
	filledBars := (value * totalBars) / 100

	progressBar := strings.Repeat("█", filledBars) + strings.Repeat("░", totalBars-filledBars)

	styleColor := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Margin(1)

	fmt.Println(styleColor.Render(
		msg + ":" + progressBar + fmt.Sprintf("%d%%", value)),
	)
}
