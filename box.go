package fui

import (
	//"fmt"
	//"strings"
	//"time"

	. "github.com/ilenker/fui/internal/calc"

	tc "github.com/gdamore/tcell/v3"
	//rw "github.com/mattn/go-runewidth"
)

type BoxType int

const (
	TerminalT BoxType = iota
	ButtonT
	PromptT
	TxtpadT
)

type UI struct {
	Layout   LayoutData
	Terms    TerminalData
	Buttons  ButtonData
	Names    []string
	Count	 int
}

// Shared fields, hot
type LayoutData struct {
	Rects     []Rect
	RectsPrev []Rect
	Types     []BoxType
}

func (ui *UI) AddTerminal(name string) int {
	id := ui.Count
	ui.Count++
	ui.Layout.Rects[id] = Rect{ X: 5, Y: 5, W: 10, H: 10, }
	ui.Layout.RectsPrev[id] = Rect{ X: 5, Y: 5, W: 10, H: 10, }
	ui.Layout.Types[id] = TerminalT
	ui.Names[id] = name
	ui.Terms.Buffers[id] = make([]byte, 0, 1024)
	ui.Terms.Lines[id] = make([]Span, 0, 128)
	ui.Terms.Views[id] = 0
	return id
}

var boxDefault = [7]rune{'┌', '┐', '└', '┘', '│', '─', '○'}
var boxFocused = [7]rune{'┏', '┓', '┗', '┛', '┃', '━', '●'}

func (ui *UI) DrawBorder(id int) {
	rect := ui.Layout.Rects[id]
	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	var style tc.Style
	var rs *[7]rune

	switch id {
	case focusedIdx:
		style = stBorderFocused
		rs = &boxFocused
	default:
		style = stTerminalBorder
		rs = &boxDefault
	}
	// Sides
	var lenLines int
	if len(ui.Terms.Buffers[id]) == 0 {
		lenLines = 0
	} else {
		lenLines = ui.Layout.Rects[id].W / len(ui.Terms.Buffers[id])
	}

	f := ILerp(0, lenLines, float64(ui.Terms.Views[id]))
	sliderPos := int(Lerp(0, h, f))
	for i := range h {
		scr.SetContent(x-1, y+i, rs[4], nil, style)
		scr.SetContent(x+w, y+i, rs[4], nil, style)
	}
	if ui.Layout.Types[id] != ButtonT {
		scr.SetContent(x+w, y+sliderPos, '█', nil, style)
	}
	// Top
	for i := range w {
		scr.SetContent(x+i, y-1, rs[5], nil, style)
	}
	// data.ttom
	for i := range w {
		scr.SetContent(x+i, y+h, rs[5], nil, style)
	}
	// Corners
	switch ui.Layout.Types[id] {
	case ButtonT:
		scr.SetContent(x+w, y+h, rs[3], nil, style) // Bottom Right without handle
	default:
		scr.SetContent(x+w, y+h, rs[6], nil, style) // Bottom Right with handle
	}
	scr.SetContent(x-1, y-1, rs[6], nil, style) // Top Left with handle
	scr.SetContent(x+w, y-1, rs[1], nil, style) // Top Right
	scr.SetContent(x-1, y+h, rs[2], nil, style) // Bottom Left

	// Labels
	switch ui.Layout.Types[id] {
	case ButtonT:
		// Do nothing
	case TxtpadT:
		for i, r := range ui.Names[id] {
			if i >= w {
				scr.SetContent(x+i, y-1, '…', nil, stDimmed)
				return
			}
			scr.SetContent(x+i, y-1, rune(r), nil, stDimmed)
		}
	default:
		for i, r := range ui.Names[id] {
			if i >= w {
				scr.SetContent(x+i, y-1, '…', nil, stDimmed)
				return
			}
			scr.SetContent(x+i, y-1, rune(r), nil, stDef)
		}
	}
}

type Rect struct {
	X, Y, W, H int
}

type Span struct {
	Start int
	End   int
}

type TerminalData struct {
	Buffers  [][]byte
	Lines    [][]Span
	Views    []int
	CursorXs []int
	CursorYs []int
}

func (td *TerminalData) WriteToBuffer(id int, s string) {
	td.Buffers[id] = append(td.Buffers[id], []byte(s)...)
}

func DrawTerminal(id int, ui *UI) {
	// Don't try rendering if the width is zero
	ui.DrawBorder(id)
	r := ui.Layout.Rects[id]
	if r.W <= 0 {
		return
	}
	if len(ui.Terms.Buffers[id]) == 0 {
		return
	}
	y := r.Y
	yLimit := r.Y + r.H - 1
	ts := &ui.Terms
	for i := ts.Views[id]; i < len(ts.Lines[id]); i++ {
		l := ts.Lines[id][i]
		line := ts.Buffers[id][l.Start : l.End]
		for i, c := range line {
			scr.SetContent(r.X+i, y, rune(c), nil, stDef)
		}
		y++
		if y > yLimit {
			break
		}
	}
}

func ReflowLines(lines *[]Span, buf []byte, width int) {
	*lines = (*lines)[:0]
	lineStart := 0
	prevWithinLimit := 0
	for i := range buf {
		if buf[i] == '\n' {
			*lines = append(*lines, Span{lineStart, i})
			lineStart = i+1
			prevWithinLimit = lineStart
			continue
		}
		currentLineWidth := 1+i-lineStart
		if buf[i] == ' ' {
			// Confirm if current line width is within limit
			if currentLineWidth <= width {
				prevWithinLimit = i
			} else {
				// Does not fit, use prevWithinLimit as end
				*lines = append(*lines, Span{lineStart, prevWithinLimit})
				lineStart = prevWithinLimit+1
				prevWithinLimit++
			}
			continue
		}
		if currentLineWidth >= width {
			if prevWithinLimit == lineStart {
				prevWithinLimit = i
			}
			*lines = append(*lines, Span{lineStart, prevWithinLimit})
			lineStart = prevWithinLimit+1
			prevWithinLimit++
		}
	}
	*lines = append(*lines, Span{lineStart, len(buf)})
}

// Not sure about this yet
type ButtonData struct {
	UserFunctions []uint8
}

type Box struct {
	Toks    []string // Terminal, Prompt?
	Lines   []Span   // Terminal
	BoxType BoxType  // All x
	Name    string   // All x
	View    int      // Terminal, Prompt?
	X, Y    int      // All x
	W, H    int      // All x
	cursor  struct { // Terminal, Prompt?
		X, Y int
	}
	id   int         // ???
	prev struct {    // All x
		X, Y int
		W, H int
	}
}

type Zone uint8
const (
	ZoneNone Zone = iota
	ZoneTopL
	ZoneBotR
	ZoneFaceR
	ZoneInside
)
func (d Zone) String() string {
	return [...]string{"NO", "TL", "BR", "FR", "IN"}[d]
}

func (ui *UI) UpdateMouseState() {
	mx, my := Mouse.X, Mouse.Y
	rects := ui.Layout.Rects

	// Find id that is colliding with mouse
	for i := ui.Count; i >= 0; i-- {
		zone := rects[i].WhichZone(mx, my)
		if zone != ZoneNone {
			Mouse.HotID = i
			Mouse.HotZone = zone
			return
		}
	}
	Mouse.HotID = -1
	Mouse.HotZone = ZoneNone
}

func (ui *UI) ApplyMouseState() {
	if Mouse.ActID == -1 &&
	   Mouse.HotID == -1 {
		return
	}
	switch Mouse.Mask {
	case tc.Button1:
		// A Click Event
		if Mouse.PrevMask != tc.Button1 {
			Mouse.ActID = Mouse.HotID
			Mouse.ActZone = Mouse.HotZone
			return
		}
		if Mouse.PrevMask == tc.Button1 {
		// A Drag Event
			switch Mouse.ActZone {
			case ZoneTopL:
				rect := &ui.Layout.Rects[Mouse.ActID]
				rect.X = Mouse.X+1
				rect.Y = Mouse.Y+1
			case ZoneBotR:
				rect := &ui.Layout.Rects[Mouse.ActID]
				prevW := rect.W
				rect.W = Mouse.X - rect.X
				rect.H = Mouse.Y - rect.Y
				if rect.W < 0 {
					rect.W = 0
				}
				if rect.H < 0 {
					rect.H = 0
				}
				if rect.W != prevW {
					prevLenLines := len(ui.Terms.Lines[Mouse.ActID])
					prevView := ui.Terms.Views[Mouse.ActID]
					ReflowLines(&ui.Terms.Lines[Mouse.ActID], ui.Terms.Buffers[Mouse.ActID], rect.W)
					if prevLenLines == 0 {
						return
					}
					ratio := float64(prevView) / float64(prevLenLines)
					adjustedView := float64(len(ui.Terms.Lines[Mouse.ActID])) * ratio
					ui.Terms.Views[Mouse.ActID] = int(adjustedView)
				}
			}
		}
	case tc.WheelDown:
		switch Mouse.HotZone {
		case ZoneInside:
			if ui.Terms.Views[Mouse.HotID] < len(ui.Terms.Lines[Mouse.HotID]) {
				ui.Terms.Views[Mouse.HotID]++
			}
		}
	case tc.WheelUp:
		switch Mouse.HotZone {
		case ZoneInside:
			if ui.Terms.Views[Mouse.HotID] > 0 {
				ui.Terms.Views[Mouse.HotID]--
			}
		}
	case tc.ButtonNone:
		Mouse.ActID = -1
		Mouse.ActZone = ZoneNone
	}
}

func (r *Rect) WhichZone(mX, mY int) Zone {
	rX1 := r.X-1
	rY1 := r.Y-1
	rX2 := r.X+r.W
	rY2 := r.Y+r.H
	switch {
	case mX < rX1 || mY < rY1:
		return ZoneNone
	case mX > rX2 || mY > rY2:
		return ZoneNone
	case (mX == r.X || mX == rX1) && mY == rY1:
		return ZoneTopL
	case mX == rX2 && mY == rY2:
		return ZoneBotR
	default:
	return ZoneInside
	}
}
