package prober

import (
	"fmt"
	"time"
	"reflect"
	"github.com/ilenker/prober/internal/calc"

	tc "github.com/gdamore/tcell/v3"
)


type BufferType int

const (
	Text BufferType = iota
	Button
)

type Buffer struct {
	Type	BufferType
	Toks    []string
	Lines   []struct{
		Start int
		End   int
	}
	View    int
	X, Y    int
	W, H    int
	ID 		int
	Cursor  struct{X, Y int}
	Prev struct{
		X, Y int
		W, H int
	}
	Style	tc.Style
	Write			func(s string)
	Draw			func()
	OnMouseEvent	func(*tc.EventMouse)
	OnClick			func()
	OnUpdate		func()
	Watch			struct{current any; previous any}
}

func (b *Buffer) textWrite(s string) {
	//toks := strings.SplitAfter(s, " ")
	toks := tokenize(s)
	b.Toks = append(b.Toks, toks...)

	// Just reflow everything everytime for now
	b.reflowLines()
	b.textDraw()
}


func (b *Buffer) buttonWrite(s string) {
	toks := tokenize(s)
	b.Toks = append(b.Toks, toks...)

	// Just reflow everything everytime for now
	b.reflowLines()
	b.buttonDraw()
}


// Stages the entire buffer region.
func (b *Buffer) textDraw() {
	// Don't try rendering if the width is zero
	if b.W <= 0 {
		return
	}
	if len(b.Toks) == 0 {
		return
	}
	for i := b.View; i < len(b.Lines); i++ {
		if i-b.View >= b.H { break }
		span := b.Lines[i]
		line := ""
		for t := span.Start;
			t <= span.End;
			t++ {
			line += b.Toks[t]
		}
		for x, r := range line {
			y := i - b.View
			scr.SetContent(b.X+x, b.Y + y, r, nil, b.Style)
		}
	}
}


func (b *Buffer) reflowLines() {
	if len(b.Toks) == 0 {
		return
	}
	b.Lines = b.Lines[:0]
	chars := 0
	start := 0

	for i, tok := range b.Toks {
		l := len(tok)
		if l == 0 {
			continue
		}
		if (chars + l > b.W && i > start) {
			newRange := struct{Start, End int}{start, i - 1}
			b.Lines = append(b.Lines, newRange)
			start = i
			chars = 0
		}
		chars += l

		if (tok[l-1] == '\n') {
			newRange := struct{Start, End int}{start, i}
			b.Lines = append(b.Lines, newRange)
			start = i + 1
			chars = 0
		}
	}
	// Trailing line
	if start < len(b.Toks) {
		newRange := struct{Start, End int}{start, len(b.Toks) - 1}
		b.Lines = append(b.Lines, newRange)
	}
}


func (b *Buffer) buttonDraw() {
	// Don't try rendering if the width is zero
	if b.W <= 0 {
		return
	}
	if len(b.Toks) == 0 {
		return
	}
	for i := b.View; i < len(b.Lines); i++ {
		if i > b.H { break }
		span := b.Lines[i]
		line := ""
		for t := span.Start;
			t <= span.End;
			t++ {
			line += b.Toks[t]
		}
		for x, r := range line {
			y := i - b.View
			scr.SetContent(b.X+x, b.Y + y, r, nil, b.Style)
		}
	}
}


func (b *Buffer) textOnMouseEvent(ev *tc.EventMouse) {
	x, y    := ev.Position()
	buttons := ev.Buttons()
	mouse.Buttons = buttons

	defer func() {
		mouse.Prev.Buttons = buttons
		redraw = true
		b.reflowLines()
	}()

	switch buttons {
	case tc.Button1:
		// Move
		if b.onTopLeft(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.ID
			} else 
			if focusedIdx != b.ID {
				return
			}
			b.X, b.Y = x, y
		} else
		// Resize
		if b.onBottomRight(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.ID
			} else
			if focusedIdx != b.ID {
				return
			}
			b.W, b.H = x-b.X, y-b.Y
		} else
		if b.onScrollRegion(x, y) {
			b.updateSlider(x, y)
		} else
		if focusedIdx == b.ID {
			focusedIdx = -1
		}
	// Scroll up and down
	case tc.WheelUp:
		if b.onScrollRegion(x, y) {
			b.View = max(b.View-1, 0)
		}
	case tc.WheelDown:
		if b.onScrollRegion(x, y) {
			b.View = min(b.View+1, len(b.Lines)-1)
		}
	default:
	}
}

func (b *Buffer) buttonOnMouseEvent(ev *tc.EventMouse) {
	x, y    := ev.Position()
	buttons := ev.Buttons()
	mouse.Buttons = buttons

	defer func() {
		redraw = true
	}()

	switch buttons {
	case tc.Button1:
		// Move
		if b.onTopLeft(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.ID
			} else 
			if focusedIdx != b.ID {
				return
			}
			b.X, b.Y = x, y
		} else
		if b.inMe(x, y) {
			b.Flash()
			b.OnClick()
		} else
		if focusedIdx == b.ID {
			focusedIdx = -1
		}
	default:
	}
}

func (b *Buffer) Flash() {
	b.Style = b.Style.Reverse(true)
	go func() {
		t := time.NewTimer(time.Millisecond * 100)
		<-t.C
		b.Style = b.Style.Reverse(false)
		redraw = true
	}()
}


func (b *Buffer) onScrollRegion(x, y int) bool {
	if x == b.X + b.W &&
	   y >= b.Y && y < b.Y + b.H {
		return true
	}
	return false
}

func (b *Buffer) updateSlider(x, y int) {
	f := calc.ILerp(b.Y, b.Y+b.H, float64(y))
	f = calc.Clamp(f, 0, 1)

	// TODO: Redo
	//b.ViewIdx = int(calc.Lerp(0, len(b.Lines)-1, f))
}

func (b *Buffer) onTopLeft(x2, y2, dist int) bool {
	xDist := max(b.X, x2) - min(b.X, x2)
	yDist := max(b.Y, y2) - min(b.Y, y2)
	return yDist <= dist && xDist <= dist
}

func (b *Buffer) onBottomRight(x2, y2, dist int) bool {
	xDist := max(b.X + b.W, x2) - min(b.X + b.W, x2)
	yDist := max(b.Y + b.H, y2) - min(b.Y + b.H, y2)
	return yDist <= dist && xDist <= dist
}

func (b *Buffer) inMe(x, y int) bool {
	inX := x >= b.X && x <= b.X + b.W
	inY := y >= b.Y && y <= b.Y + b.H
	return inX && inY
}

func (b *Buffer) nextEOW(from int) int {
	return 0
}


func (b *Buffer) Clear() {
	X, Y, W, H := b.X, b.Y, b.W, b.H
	for y := Y;
		y < Y+H;
		y++ {
		for x := X;
			x < X+W;
			x++ {
			scr.SetContent(
				x, y,
				' ',
				nil,
				stDef)
		}
	}
}


var boxThin  = [6]rune{'┌', '┐', '└', '┘', '│', '─'}
var boxThick = [6]rune{'┏', '┓', '┗', '┛', '┃', '━'}

func (b *Buffer) box() {
	x, y, w, h := b.X, b.Y, b.W, b.H
	rs := boxThin[:]
	style := b.Style

	if b.ID == focusedIdx {
		style = b.Style.Foreground(tc.ColorRed)
		rs = boxThick[:]
	}

	// Sides
	// TODO: Redo
	//f := calc.ILerp(0, len(b.Lines), float64(b.ViewIdx))
	//sliderPos := int(calc.Lerp(0, h, f))
	for i := range h {
		scr.SetContent(x - 1, y + i, rs[4], nil, style)
		scr.SetContent(x+w  , y + i, rs[4], nil, style)
	}
	//scr.SetContent(x+w  , y + sliderPos, '█', nil, style)
	// Top/Bottom
	for i := range w {
		scr.SetContent(x + i, y - 1, rs[5], nil, style)
		scr.SetContent(x + i, y+h  , rs[5], nil, style)
	}
	// Corners
	scr.SetContent(x - 1, y - 1, rs[0], nil, style)
	scr.SetContent(x+w  , y - 1, rs[1], nil, style)
	scr.SetContent(x - 1, y+h  , rs[2], nil, style)
	scr.SetContent(x+w  , y+h  , rs[3], nil, style)
}


func (b *Buffer) clearBox() {
	x, y, w, h := b.X, b.Y, b.W, b.H
	// Sides
	style := tc.StyleDefault
	for i := range h {
		scr.SetContent(x - 1, y + i, ' ', nil, style)
		scr.SetContent(x+w  , y + i, ' ', nil, style)
	}
	// Top/Bottom
	for i := range w {
		scr.SetContent(x + i, y - 1, ' ', nil, style)
		scr.SetContent(x + i, y+h  , ' ', nil, style)
	}
	// Corners
	scr.SetContent(x - 1, y - 1, ' ', nil, style)
	scr.SetContent(x+w  , y - 1, ' ', nil, style)
	scr.SetContent(x - 1, y+h  , ' ', nil, style)
	scr.SetContent(x+w  , y+h  , ' ', nil, style)
}


// Struct makers -------------------------------------
func NewTerminal(x, y, w, h int) *Buffer {
	toks  := make([]string, 0)
	lines := make([]struct{Start int; End int}, 1)
	b := &Buffer{
		Type: Text,
		Toks: toks,
		Lines: lines,
		X: x,
		Y: y,
		W: w,
		H: h,
		ID: id,
	}
	b.Draw = b.textDraw
	b.Write = b.textWrite
	b.OnMouseEvent = b.textOnMouseEvent
	b.OnUpdate = func(){}
	b.box()
	buffers = append(buffers, b)
	id++
	return b
}

func (b *Buffer) flash() {
}

func NewButton(x, y int, label string, onClick func()) *Buffer {
	toks  := make([]string, 1)
	toks[0] = label
	lines := make([]struct{Start int; End int}, 1)
	b := &Buffer{
		Type: Text,
		Toks: toks,
		Lines: lines,
		X: x,
		Y: y,
		W: len(label),
		H: 1,
		ID: id,
	}
	b.Draw = b.buttonDraw
	b.Write = b.buttonWrite
	b.OnMouseEvent = b.buttonOnMouseEvent
	b.OnClick = onClick
	b.OnUpdate = func(){}
	buffers = append(buffers, b)
	b.box()
	id++
	return b
}

func NewWatcher(x, y int, label string, v any) *Buffer {
	toks      := make([]string, 1)
	lines     := make([]struct{Start int; End int}, 1)
	b := &Buffer{
		Type: Button,
		Toks: toks,
		Lines: lines,
		X: x,
		Y: y,
		H: 3,
		ID: id,
	}
	b.Watch.current = v
	b.Draw = b.buttonDraw
	b.Write = b.buttonWrite
	b.OnMouseEvent = b.buttonOnMouseEvent
	b.OnUpdate = func(){
		v := reflect.ValueOf(b.Watch.current)
		if v.Elem() != b.Watch.previous {
			b.Toks[0] = fmt.Sprintf("%s: %v", label, v.Elem())
			b.Watch.previous = v.Elem().Interface()
			b.Draw()
		}
	}
	b.OnUpdate()
	b.W = len(b.Toks[0])
	b.OnClick = func(){}
	buffers = append(buffers, b)
	b.box()
	id++
	return b
}

func tokenize(s string) []string {
	toks := make([]string, 0)

	// Scan over every character in the string
	start := 0
	for i := range s {
		// If we reach end and it's not a newline
		// we may need to add to it later
		// perhaps a flag for "unterminated last token"
		if i == len(s)-1 {
			tok := s[start: i+1]
			toks = append(toks, tok)
			break
		}
		// Look for spaces or '\n'
		// Every excess space gets its own token right now.
		// Deal with it later
		if s[i] == ' ' ||
		   s[i] == '\n' {
			tok := s[start: i+1]
			toks = append(toks, tok)
			start = i+1
		}
	}
	return toks
}
//  0123456789ABCD EF
// [apple crumble]
/*
	[
		"apple "
		"crumble\n"
		"topple "
		"whipple\n"
		"\n"
		"cripplenipple"
	]

apple crumble
topple whipple

cripplenipple
*/
