package fui

import (
	"os"
	tc "github.com/gdamore/tcell/v3"
)

var colortermOG string

// Color ("style") definitions
var (
	stDimmed tc.Style
	stBody   tc.Style
	stTerminalBorder tc.Style
	stWatcherBorder    tc.Style
	stButtonBorder   tc.Style
)

func initStyles() {
	stDimmed = tc.StyleDefault.Foreground(tc.NewRGBColor(100, 100, 100))
	d := int32(2)
	stTerminalBorder = tc.StyleDefault.Foreground(tc.NewRGBColor(  90/d, 225/d, 180/d))
	stWatcherBorder  = tc.StyleDefault.Foreground(tc.NewRGBColor( 125/d, 100/d,  80/d))
	stButtonBorder   = tc.StyleDefault.Foreground(tc.NewRGBColor( 180/d, 100/d, 180/d))
}

func setCOLORTERM() {
	colortermOG = os.Getenv("COLORTERM")
	os.Setenv("COLORTERM", "truecolor")	
}

func restoreCOLORTERM() {
	os.Setenv("COLORTERM", colortermOG)
}
