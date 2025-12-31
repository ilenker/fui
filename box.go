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
	Txtpads  TxtpadData
	Names    [][32]byte
	Count	 int
}

// Shared fields, hot
type LayoutData struct {
	Rects     []Rect
	RectsPrev []Rect
	Types     []BoxType
}

var	Names [][32]byte

func (ui *UI) AddTerminal(name string) int {
	id := ui.Count
	ui.Count++
	ui.Layout.Rects[id] = Rect{ X: 5, Y: 5, W: 10, H: 10, }
	ui.Layout.RectsPrev[id] = Rect{ X: 5, Y: 5, W: 10, H: 10, }
	ui.Layout.Types[id] = TerminalT
	b := [32]byte{}
	copy(b[:], name)
	ui.Names[id] = b
	ui.Terms.Buffers[id] = make([]byte, 0, 1024)
	ui.Terms.Views[id] = 0
	return id
}

var boxDefault = [7]rune{'┌', '┐', '└', '┘', '│', '─', '○'}
var boxFocused = [7]rune{'┏', '┓', '┗', '┛', '┃', '━', '●'}

func (ui *UI) DrawBorder(id int) {
	x, y, w, h := ui.Layout.Rects[id].X, ui.Layout.Rects[id].Y, ui.Layout.Rects[id].W, ui.Layout.Rects[id].H
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

// Specifics
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
	for y := range r.H {
		for x := range r.W {
			i := (y*r.W)+x
			if i >= len(ui.Terms.Buffers[id]) {
				return
			}
			scr.SetContent(r.X + x, r.Y + y, rune(ui.Terms.Buffers[id][i]), nil, stDef)
		}
	}
}


// Keeping this one as is for now
// Not yet ready to tackle word wrap
// with a pure byte buffer...
type TxtpadData struct {
	Tokens   [][]string
	Lines    [][]Span
	Views    []int
	CursorXs []int
	CursorYs []int
}

// Not sure about this yet
type ButtonData struct {
	UserCallbacks []uint8
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
		zone := rects[i].GetZone(mx, my)
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
				ui.Layout.Rects[Mouse.ActID].X = Mouse.X+1
				ui.Layout.Rects[Mouse.ActID].Y = Mouse.Y+1
			}
		}
	case tc.ButtonNone:
		Mouse.ActID = -1
		Mouse.ActZone = ZoneNone
	}
}

func (r *Rect) GetZone(mX, mY int) Zone {
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

func mousePosChanged() bool {
	xChange := Mouse.X != Mouse.PrevX
	yChange := Mouse.Y != Mouse.PrevY
	return xChange || yChange
}

func (r *Rect) updateSlider(x, y int) {
	f := ILerp(r.Y, r.Y+r.H, float64(y))
	f = Clamp(f, 0, 1)
	// TODO: Redo
	//b.ViewIdx = int(Lerp(0, len(b.lines)-1, f))
}

func (r *Rect) onScrollRegion(x, y int) bool {
	if x == r.X+r.W &&
		y >= r.Y && y < r.Y+r.H {
		return true
	}
	return false
}

func (r *Rect) onTopLeft(x2, y2, dist int) bool {
	xDist := Abs(r.X - 1 - x2)
	yDist := Abs(r.Y - 1 - y2)
	return yDist <= dist && xDist <= dist
}

func (r *Rect) onBottomRight(x2, y2, dist int) bool {
	xDist := Abs(r.X + r.W - x2)
	yDist := Abs(r.Y + r.H - y2)
	return yDist <= dist && xDist <= dist
}

func (r *Rect) inMe(x, y int) bool {
	inX := x >= r.X-1 && x <= r.X+r.W
	inY := y >= r.Y-1 && y <= r.Y+r.H
	return inX && inY
}

// ===============================================================================================
// ==================================Old, Broken Code Below=======================================
// ==================================Odl, Croken Bode Lebow=======================================
// ===============================================================================================

//func (b *Box) restoreLayout() {
//	if !layoutLoadOK {
//		return
//	}
//	for j, entry := range restoredLayout {
//		if entry.Name == b.Name {
//			b.X = entry.X
//			b.Y = entry.Y
//			if entry.Type != int(buttonT) {
//				b.W = entry.W
//				b.H = entry.H
//			}
//			restoredLayout[j].Name = "^=_$" + entry.Name
//			b.reflowLines()
//			break
//		}
//	}
//}
//
//// _Write -------------------------------------------------------------------------------------
//func (b *Box) textWrite(s string) {
//	toks := tokenize(s)
//	b.toks = append(b.toks, toks...)
//
//	// Just reflow everything everytime for now
//	b.reflowLines()
//	// Stick view to bottom only
//	// if already at the bottom
//	if b.view == len(b.lines)-b.H-1 {
//		b.view = len(b.lines) - b.H
//	}
//	b.textDraw()
//}
//
//func (b *Box) fieldWrite(s string) {
//	// This will always be one character
//	// until we implement copy paste
//	if len(s) == 0 {
//		return
//	}
//	toks := tokenize(s)
//	var lastTokI int
//	var lastCharI int
//
//	// If there are no tokens, add new and move on, no checks
//	if len(b.toks) == 0 {
//		b.toks = append(b.toks, toks...)
//		goto skipCheck
//	}
//
//	lastTokI = len(b.toks) - 1
//	lastCharI = len(b.toks[lastTokI]) - 1
//	// Last token was terminated - just append new tokens
//	if b.toks[lastTokI][lastCharI] == '\n' {
//		b.toks = append(b.toks, toks...)
//		// Last token unterminated - don't make new, add to it directly
//	} else {
//		b.toks[lastTokI] += toks[0]
//	}
//
//skipCheck:
//	// Just reflow everything everytime for now
//	b.reflowLines()
//	if s == "\n" {
//		b.view++
//	} else {
//		if b.view != len(b.lines) - 1 {
//			b.view = len(b.lines) - 1
//		}
//	}
//	b.textDraw()
//}
//
//func (b *Box) buttonWrite(s string) {
//	toks := tokenize(s)
//	b.toks = append(b.toks, toks...)
//	b.buttonDraw()
//}
//
//// _Draw -------------------------------------------------------------------------------------
//
//func (b *Box) textDraw() {
//	// Don't try rendering if the width is zero
//	if b.W <= 0 {
//		return
//	}
//	if len(b.toks) == 0 {
//		return
//	}
//	for i := b.view; i < len(b.lines); i++ {
//		if i-b.view >= b.H {
//			break
//		}
//		span := b.lines[i]
//		builder := strings.Builder{}
//		for t := span.start; t <= span.end; t++ {
//			builder.WriteString(b.toks[t])
//		}
//		x := 0
//		for _, r := range builder.String() {
//			y := i - b.view
//			if x >= b.W {
//				break
//			}
//			scr.SetContent(b.X+x, b.Y+y, r, nil, b.BodyStyle)
//			b.cursor.X = x
//			b.cursor.Y = y
//			if r > 255 {
//				x += rw.RuneWidth(r)
//			} else {
//				x++
//			}
//		}
//	}
//}
//
//// Border
//
//
//func (b *Box) clearBorder() {
//	x, y, w, h := b.X, b.Y, b.W, b.H
//	// Sides
//	style := tc.StyleDefault
//	for i := range h {
//		scr.SetContent(x-1, y+i, ' ', nil, style)
//		scr.SetContent(x+w, y+i, ' ', nil, style)
//	}
//	// Top/Bottom
//	for i := range w {
//		scr.SetContent(x+i, y-1, ' ', nil, style)
//		scr.SetContent(x+i, y+h, ' ', nil, style)
//	}
//	// Corners
//	scr.SetContent(x-1, y-1, ' ', nil, style)
//	scr.SetContent(x+w, y-1, ' ', nil, style)
//	scr.SetContent(x-1, y+h, ' ', nil, style)
//	scr.SetContent(x+w, y+h, ' ', nil, style)
//}
//
//func (b *Box) buttonDraw() {
//	// Don't try rendering if the width is zero
//	if b.W <= 0 {
//		return
//	}
//	x := 0
//	for _, r := range b.Name {
//		scr.SetContent(b.X+x, b.Y, r, nil, b.BodyStyle)
//		x++
//	}
//}
//
//// _Token & Line Processing --------------------------------------------------------------------------------
//
//func tokenize(s string) []string {
//	toks := make([]string, 0)
//	// Scan over every character in the string
//	start := 0
//	for i := range len(s) {
//		// If we reach end and it's not a newline
//		// we may need to add to it later
//		// perhaps a flag for "unterminated last token"
//		if i == len(s)-1 {
//			tok := s[start : i+1]
//			toks = append(toks, tok)
//			break
//		}
//		// Look for spaces or '\n'
//		// Every excess space gets its own token right now.
//		// TODO: Deal with it later
//		if s[i] == ' ' ||
//			s[i] == '\n' {
//			tok := s[start : i+1]
//			toks = append(toks, tok)
//			start = i + 1
//		}
//	}
//	return toks
//}
//
//func (b *Box) reflowLines() {
//	if len(b.toks) == 0 {
//		return
//	}
//	b.lines = b.lines[:0]
//	chars := 0
//	start := 0
//
//	for i, tok := range b.toks {
//		l := len(tok)
//		if l == 0 {
//			continue
//		}
//		if chars+l > b.W && i > start {
//			newRange := span{start, i - 1}
//			b.lines = append(b.lines, newRange)
//			start = i
//			chars = 0
//		}
//		chars += l
//		if tok[l-1] == '\n' {
//			newRange := span{start, i}
//			b.lines = append(b.lines, newRange)
//			start = i + 1
//			chars = 0
//		}
//	}
//	// Trailing line
//	if start < len(b.toks) {
//		newRange := span{start, len(b.toks) - 1}
//		b.lines = append(b.lines, newRange)
//	}
//}
//
//// _OnHot --------------------------------------------------------------------------------
//
//func (b *Box) terminalOnHot(ev *tc.EventMouse) {
//	update := func() {
//		Redraw = true
//		b.reflowLines()
//	}
//	x, y := mouse.x, mouse.y
//	moved := mousePosChanged()
//	mods := ev.Modifiers()
//	lenLines := len(b.lines)
//	switch mouse.mask {
//	case tc.Button1:
//		// Move
//		switch {
//		case b.onTopLeft(x, y, 1):
//			if focusedIdx == -1 {
//				focusedIdx = b.id
//			} else if focusedIdx != b.id {
//				return
//			}
//			if moved {
//				b.X, b.Y = x+1, y+1
//				update()
//			}
//		case b.onBottomRight(x, y, 1):
//			if focusedIdx == -1 {
//				focusedIdx = b.id
//			} else if focusedIdx != b.id {
//				return
//			}
//			update()
//			hPrev := b.H
//			b.W, b.H = x-b.X, y-b.Y
//			// Stick to the bottom of the view logic
//			if b.view > 0 &&
//				lenLines-b.view > b.H {
//				b.view -= b.H-hPrev
//			}
//		case b.onScrollRegion(x, y):
//			b.updateSlider(x, y)
//		case b.inMe(x, y):
//			if mods == tc.ModCtrl {
//				b.view = lenLines - b.H
//				if b.view < 0 {
//					b.view = 0
//				}
//			}
//		case focusedIdx == b.id:
//			focusedIdx = -1
//		}
//	// Scroll up and down
//	case tc.WheelUp:
//		if b.onScrollRegion(x, y) {
//			if mods == tc.ModCtrl {
//				b.view = max(b.view-10, 0)
//			} else {
//				b.view = max(b.view-1, 0)
//			}
//			update()
//		}
//	case tc.WheelDown:
//		if b.onScrollRegion(x, y) {
//			if mods == tc.ModCtrl {
//				b.view = min(b.view+10, lenLines-1)
//			} else {
//				b.view = min(b.view+1, lenLines-1)
//			}
//			update()
//		}
//	default:
//	}
//}
//
//func (b *Box) buttonHot(ev *tc.EventMouse) {
//	Redraw = true
//	x := mouse.x
//	y := mouse.y
//
//	switch mouse.mask {
//	case tc.Button1:
//		if focusedIdx == b.id {
//			focusedIdx = -1
//		}
//	case tc.Button2:
//		if b.onTopLeft(x, y, 1) {
//			if focusedIdx == -1 {
//				focusedIdx = b.id
//			} else if focusedIdx != b.id {
//				return
//			}
//			b.X, b.Y = x+1, y+1
//		} else if focusedIdx == b.id {
//			focusedIdx = -1
//		}
//	default:
//		if b.inMe(x, y) {
//			if mouse.prev.mask == tc.Button1 {
//				b.Flash()
//				b.OnClick(b)
//			}
//		}
//	}
//}
//
//func (b *Box) textOnHot(ev *tc.EventMouse) {
//	x, y := mouse.x, mouse.y
//	moved := mousePosChanged()
//	mods := ev.Modifiers()
//
//	update := func() {
//		b.reflowLines()
//	}
//	switch mouse.mask {
//	case tc.Button1:
//		// Move
//		if b.onTopLeft(x, y, 1) {
//			if focusedIdx == -1 {
//				focusedIdx = b.id
//			} else if focusedIdx != b.id {
//				return
//			}
//			if moved {
//				b.X, b.Y = x+1, y+1
//				update()
//			}
//		} else
//		// Resize
//		if b.onBottomRight(x, y, 1) {
//			if focusedIdx == -1 {
//				focusedIdx = b.id
//			} else if focusedIdx != b.id {
//				return
//			}
//			update()
//			b.W, b.H = x-b.X, y-b.Y
//		} else if b.onScrollRegion(x, y) {
//			b.updateSlider(x, y)
//		} else if focusedIdx == b.id {
//			focusedIdx = -1
//		}
//	// Scroll up and down
//	case tc.WheelUp:
//		if b.onScrollRegion(x, y) {
//			if mods == tc.ModCtrl {
//				b.view = max(b.view-10, 0)
//			} else {
//				b.view = max(b.view-1, 0)
//			}
//			update()
//		}
//	case tc.WheelDown:
//		if b.onScrollRegion(x, y) {
//			if mods == tc.ModCtrl {
//				b.view = min(b.view+10, len(b.lines)-1)
//			} else {
//				b.view = min(b.view+1, len(b.lines)-1)
//			}
//			update()
//		}
//	default:
//	}
//}
//
//func (b *Box) setPrevs() {
//	b.prev.X, b.prev.Y = b.X, b.Y
//	b.prev.W, b.prev.H = b.W, b.H
//}
//
//func mousePosChanged() bool {
//	xChange := mouse.x != mouse.prev.x
//	yChange := mouse.y != mouse.prev.y
//	return xChange || yChange
//}
//
//func (b *Box) updateSlider(x, y int) {
//	f := ILerp(b.Y, b.Y+b.H, float64(y))
//	f = Clamp(f, 0, 1)
//	// TODO: Redo
//	//b.ViewIdx = int(Lerp(0, len(b.lines)-1, f))
//}
//
//func (b *Box) onScrollRegion(x, y int) bool {
//	if x == b.X+b.W &&
//		y >= b.Y && y < b.Y+b.H {
//		return true
//	}
//	return false
//}
//
//func (b *Box) onTopLeft(x2, y2, dist int) bool {
//	xDist := Abs(b.X - 1 - x2)
//	yDist := Abs(b.Y - 1 - y2)
//	return yDist <= dist && xDist <= dist
//}
//
//func (b *Box) onBottomRight(x2, y2, dist int) bool {
//	xDist := Abs(b.X + b.W - x2)
//	yDist := Abs(b.Y + b.H - y2)
//	return yDist <= dist && xDist <= dist
//}
//
//func (b *Box) inMe(x, y int) bool {
//	inX := x >= b.X-1 && x <= b.X+b.W
//	inY := y >= b.Y-1 && y <= b.Y+b.H
//	return inX && inY
//}
//
//// =============================== Box _Types ===============================
//
///*
//Spawn a new, basic terminal box.
//
//Returns a pointer to the new box for calling it's methods. Example:
//
//	term := fui.Terminal("Ttyvm")
//	term.Println("Hello world!")
//	term.Clear()
//*/
//func Terminal(name string) *Box {
//	toks := make([]string, 0, 64)
//	lines := make([]span, 0, 64)
//	b := &Box{
//		Name:        name,
//		boxType:     terminalT,
//		toks:        toks,
//		lines:       lines,
//		X:           nextTerminalPos.X,
//		Y:           nextTerminalPos.Y,
//		W:           15,
//		H:           15,
//		id:          nextID,
//		BorderStyle: stTerminalBorder,
//		LabelStyle:  stTerminalLabel,
//	}
//	b.Draw = b.textDraw
//	b.Write = b.textWrite
//	b.OnHot = b.terminalOnHot
//	b.OnUpdate = func() {}
//
//	boxes = append(boxes, b)
//	nextID++
//	nextTerminalPos.X = b.X + b.W + 2
//	b.restoreLayout()
//	return b
//}
//
//// TODO: docs
//func Pad(name string, text string) *Box {
//	toks := make([]string, 0, 64)
//	lines := make([]span, 0, 64)
//	b := &Box{
//		Name:        name,
//		boxType:     padT,
//		toks:        toks,
//		lines:       lines,
//		X:           nextTerminalPos.X,
//		Y:           nextTerminalPos.Y,
//		W:           15,
//		H:           15,
//		id:          nextID,
//		BorderStyle: stTextBorder,
//		LabelStyle:  stTextLabel,
//	}
//	b.Draw = b.textDraw
//	b.Write = b.textWrite
//	b.OnHot = b.textOnHot
//	b.OnUpdate = func() {}
//	b.Write(text)
//
//	boxes = append(boxes, b)
//	nextID++
//	nextTerminalPos.X = b.X + b.W + 2
//	b.restoreLayout()
//	return b
//}
//
///*
//Spawn a button that executes a function "onClick". For example:
//
//	x := 0
//	fui.Button("+1", func(b *fui.Box) {
//		x++
//	})
//*/
//func Button(label string, onClick func(self *Box)) *Box {
//	toks := make([]string, 1)
//	lines := make([]span, 1)
//	b := &Box{
//		Name:        label,
//		boxType:     buttonT,
//		toks:        toks,
//		lines:       lines,
//		X:           nextButtonPos.X,
//		Y:           nextButtonPos.Y,
//		W:           rw.StringWidth(label),
//		H:           1,
//		id:          nextID,
//		BorderStyle: stButtonBorder,
//		LabelStyle:  stButtonLabel,
//	}
//
//	b.Draw = b.buttonDraw
//	b.Write = b.buttonWrite
//	b.OnHot = b.buttonHot
//	b.OnClick = onClick
//	b.OnUpdate = func() {}
//
//	boxes = append(boxes, b)
//	nextID++
//	nextButtonPos.X = b.X + b.W + 2
//	b.restoreLayout()
//	return b
//}
//
///*
//Spawn a text field that executes a function "onCR" (enter / carriage return / newline). For example:
//
//	prompt := fui.Field("$", func(b *fui.Box) {
//		terminal.Println(prompt.Line(-1))
//	})
//*/
//func Prompt(name string, onCR func(self *Box)) *Box {
//	toks := make([]string, 64)
//	lines := make([]span, 64)
//	b := &Box{
//		Name:        name,
//		boxType:     promptT,
//		toks:        toks,
//		lines:       lines,
//		X:           nextButtonPos.X,
//		Y:           nextButtonPos.Y,
//		W:           15,
//		H:           1,
//		id:          nextID,
//		BorderStyle: stButtonBorder,
//		LabelStyle:  stFieldLabel,
//	}
//	b.Draw = b.textDraw
//	b.Write = b.fieldWrite
//	b.OnHot = b.terminalOnHot
//	b.OnUpdate = func() {}
//
//	boxes = append(boxes, b)
//	nextID++
//	nextButtonPos.X = b.X + b.W + 2
//	b.restoreLayout()
//	return b
//}
//
//// ================================ Box _Methods ===============================
//
///*
//Make the box invert colors for 100ms
//*/
//func (b *Box) Flash() {
//	b.BodyStyle = b.BodyStyle.Reverse(true)
//	Redraw = true
//	go func() {
//		t := time.NewTimer(time.Millisecond * 100)
//		<-t.C
//		b.BodyStyle = b.BodyStyle.Reverse(false)
//		Redraw = true
//	}()
//}
//
//func (b *Box) FlashLabel() {
//	b.LabelStyle = b.LabelStyle.Reverse(true)
//	Redraw = true
//	go func() {
//		t := time.NewTimer(time.Millisecond * 50)
//		<-t.C
//		b.LabelStyle = b.LabelStyle.Reverse(false)
//		Redraw = true
//	}()
//}
//
//func (b *Box) Backspace() {
//	if len(b.toks) == 0 {
//		return
//	}
//	s := b.toks[len(b.toks)-1]
//	if len(s) == 0 {
//		return
//	}
//	runes := []rune(s)
//	// Remove the last rune (independent of width)
//	if len(runes) == 0 {
//		return
//	}
//	scr.SetContent(b.X+b.cursor.X, b.Y+b.cursor.Y, ' ', nil, b.BodyStyle)
//	b.cursor.X = max(0, b.cursor.X-1)
//	runes = runes[:len(runes)-1]
//	b.toks[len(b.toks)-1] = string(runes)
//}
//
//// Just a wrapper for fmt.Println functionality
//func (b *Box) Println(s string) {
//	// TODO: Docs
//	if s == "" {
//		return
//	}
//	toks := tokenize(s)
//	b.toks = append(b.toks, toks...)
//	b.toks[len(b.toks)-1] += "\n"
//
//	// Just reflow everything everytime for now
//	b.reflowLines()
//	b.textDraw()
//}
//
//// Just a wrapper for fmt.Printf using fmt.Sprintf
//func (b *Box) Printf(format string, a ...any) {
//    s := fmt.Sprintf(format, a...)
//    b.Write(s)
//}
//
///* Clear makes a new line and sets the view to the bottom. */
//func (b *Box) Clear() {
//	b.Write("\n\n")
//	b.view = len(b.lines)
//	for x := b.X; x < b.X+b.W; x++ {
//		for y := b.Y; y < b.Y+b.H; y++ {
//			scr.SetContent(x, y, ' ', nil, stDimmed)
//		}
//	}
//}
//
///* Returns the i-th line from a box as a string. Pass in -1 to get the newest line. */
//func (b *Box) Line(i int) string {
//	// TODO: access arbitrary lines
//	if len(b.lines) == 0 ||
//		len(b.toks) == 0 {
//		return ""
//	}
//	if i != -1 {
//		return ""
//	}
//	line := ""
//	span := b.lines[len(b.lines)-1]
//	for i := span.start; i <= span.end; i++ {
//		line += b.toks[i]
//	}
//	return line
//}
//
//func (b *Box) Reset() {
//	b.view = 0
//	clear(b.lines)
//	clear(b.toks)
//}
