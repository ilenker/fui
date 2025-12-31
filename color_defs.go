package fui

import (
	"os"

	tc "github.com/gdamore/tcell/v3"
)

var colortermOG string

var d int32 = 2

// Color (style) definitions
var (
	stDef            = tc.StyleDefault
	stDimmed         = tc.StyleDefault.Foreground(tc.NewRGBColor(100, 100, 100))
	stTerminalBorder = tc.StyleDefault.Foreground(tc.NewRGBColor(90/d, 225/d, 180/d))
	stTextBorder     = tc.StyleDefault.Foreground(tc.NewRGBColor(90/d, 90/d, 180/d))
	stWatcherBorder  = tc.StyleDefault.Foreground(tc.NewRGBColor(125/d, 100/d, 80/d))
	stButtonBorder   = tc.StyleDefault.Foreground(tc.NewRGBColor(180/d, 100/d, 180/d))
	stTerminalLabel  = tc.StyleDefault.Foreground(tc.NewRGBColor(90, 225, 180))
	stWatcherLabel   = tc.StyleDefault.Foreground(tc.NewRGBColor(125, 100, 80))
	stTextLabel      = tc.StyleDefault.Foreground(tc.NewRGBColor(90, 90, 180))
	stButtonLabel    = tc.StyleDefault.Foreground(tc.NewRGBColor(180, 100, 180))
	stFieldLabel     = tc.StyleDefault.Foreground(tc.NewRGBColor(180, 180, 100))
	stBorderFocused  = tc.StyleDefault.Foreground(tc.NewRGBColor(200,   0,   0)).Bold(true)
)

func setCOLORTERM() {
	colortermOG = os.Getenv("COLORTERM")
	os.Setenv("COLORTERM", "truecolor")
}

func restoreCOLORTERM() {
	os.Setenv("COLORTERM", colortermOG)
}
