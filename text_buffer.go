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
	terminalT BufferType = iota
	buttonT
	watcherT
)

type span struct {
	Start int
	End   int
}

type Buffer struct {
	Toks    []string
	Lines   []span
	bufferType	BufferType
	Name	string
	view    int
	X, Y    int
	W, H    int
	ID 		int
	Prev struct{
		X, Y int
		W, H int
	}
	BorderStyle	tc.Style
	BodyStyle	tc.Style
	Write		func(s string)
	Draw		func()
	OnHot		func(*tc.EventMouse)
	OnClick		func()
	OnUpdate	func()
	Watch		struct{current any; previous any}
}

func (b *Buffer) textWrite(s string) {
	if !active {
		fmt.Println(s)
		return
	}
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
	for i := b.view; i < len(b.Lines); i++ {
		if i-b.view >= b.H { break }
		span := b.Lines[i]
		line := ""
		for t := span.Start;
			t <= span.End;
			t++ {
			line += b.Toks[t]
		}
		for x, r := range line {
			y := i - b.view
			scr.SetContent(b.X+x, b.Y + y, r, nil, b.BodyStyle)
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
	for i := b.view; i < len(b.Lines); i++ {
		if i > b.H { break }
		span := b.Lines[i]
		line := ""
		for t := span.Start;
			t <= span.End;
			t++ {
			line += b.Toks[t]
		}
		for x, r := range line {
			y := i - b.view
			scr.SetContent(b.X+x, b.Y + y, r, nil, b.BodyStyle)
		}
	}
}


func (b *Buffer) textOnHot(ev *tc.EventMouse) {
	x, y := mouse.x, mouse.y
	moved := mousePosChanged()
	mods  := ev.Modifiers()

	update := func() {
		redraw = true
		b.reflowLines()
	}

	switch mouse.mask {
	case tc.Button1:
		// Move
		if b.onTopLeft(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.ID
			} else 
			if focusedIdx != b.ID {
				return
			}
			if moved {
				b.X, b.Y = x+1, y+1
				update()
			}
		} else
		// Resize
		if b.onBottomRight(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.ID
			} else
			if focusedIdx != b.ID {
				return
			}
			update()
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
			if mods == tc.ModCtrl {
				b.view = max(b.view-10, 0)
			} else {
				b.view = max(b.view-1, 0)
			}
			update()
		}
	case tc.WheelDown:
		if b.onScrollRegion(x, y) {
			if mods == tc.ModCtrl {
				b.view = min(b.view+10, len(b.Lines)-1)
			} else {
				b.view = min(b.view+1, len(b.Lines)-1)
			}
			update()
		}
	default:
	}
}

func (b *Buffer) buttonHot(ev *tc.EventMouse) {
	x := mouse.x
	y := mouse.y

	defer func() {
		redraw = true
	}()

	switch mouse.mask {
	case tc.Button1:
		if focusedIdx == b.ID {
			focusedIdx = -1
		}
	case tc.Button2:
		if b.onTopLeft(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.ID
			} else 
			if focusedIdx != b.ID {
				return
			}
			b.X, b.Y = x+1, y+1
		} else
		if focusedIdx == b.ID {
			focusedIdx = -1
		}
	default:
		if b.inMe(x, y) {
			if mouse.prev.mask == tc.Button1 {
				b.Flash()
				b.OnClick()
			}
		}
	}
}

func (b *Buffer) Flash() {
	b.BodyStyle = b.BodyStyle.Reverse(true)
	go func() {
		t := time.NewTimer(time.Millisecond * 100)
		<-t.C
		b.BodyStyle = b.BodyStyle.Reverse(false)
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
	xDist := calc.Abs(b.X-1 - x2)
	yDist := calc.Abs(b.Y-1 - y2)
	return yDist <= dist && xDist <= dist
}

func (b *Buffer) onBottomRight(x2, y2, dist int) bool {
	xDist := calc.Abs(b.X + b.W - x2)
	yDist := calc.Abs(b.Y + b.H - y2)
	return yDist <= dist && xDist <= dist
}

func (b *Buffer) inMe(x, y int) bool {
	inX := x >= b.X-1 && x <= b.X + b.W
	inY := y >= b.Y-1 && y <= b.Y + b.H
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


var boxThin  = [6]rune{'○', '┐', '└', '○', '│', '─'}
var boxThick = [6]rune{'●', '┓', '┗', '●', '┃', '━'}

func (b *Buffer) box() {
	x, y, w, h := b.X, b.Y, b.W, b.H
	rs := boxThin[:]
	style := b.BorderStyle

	if b.ID == focusedIdx {
		style = b.BorderStyle.Foreground(tc.ColorRed)
		rs = boxThick[:]
	}

	// Sides
	f := calc.ILerp(0, len(b.Lines), float64(b.view))
	sliderPos := int(calc.Lerp(0, h, f))
	for i := range h {
		scr.SetContent(x - 1, y + i, rs[4], nil, style)
		scr.SetContent(x+w  , y + i, rs[4], nil, style)
	}
	if b.bufferType != buttonT {
		scr.SetContent(x+w  , y + sliderPos, '█', nil, style)
	}
	// Top
	for i := range w {
		scr.SetContent(x + i, y - 1, rs[5], nil, style)
	}

	switch b.bufferType {
	case watcherT:
		last := len(b.Toks)-1
		i := 0
		for _, r := range b.Name {
			scr.SetContent(x+i, y-1, r, nil, b.BodyStyle); i++
		}
		scr.SetContent(x+i, y-1, '[', nil, stDimmed); i++
		for _, r := range b.Toks[last] {
			if i >= b.W { break }
			scr.SetContent(x+i, y-1, r, nil, b.BodyStyle); i++
		}
		scr.SetContent(x+i, y-1, ']', nil, stDimmed)
	default:
		for i, r := range b.Name {
			scr.SetContent(x+i, y-1, r, nil, b.BodyStyle)
		}
	}


	// Bottom
	for i := range w {
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


// Returns a pointer you can use to call
// buffer.Write()
func NewTerminal(name string) *Buffer {
	toks  := make([]string, 0)
	lines := make([]span, 1)
	b := &Buffer{
		Name: name,
		bufferType: terminalT,
		Toks: toks,
		Lines: lines,
		X: nextTerminalPos.X,
		Y: nextTerminalPos.Y,
		W: 15,
		H: 15,
		ID: newID,
		BorderStyle: stTerminalBorder,
	}
	b.Draw = b.textDraw
	b.Write = b.textWrite
	b.OnHot = b.textOnHot
	b.OnUpdate = func(){}
	b.box()
	buffers = append(buffers, b)
	newID++
	nextTerminalPos.X = b.X + b.W + 2
	return b
}


func NewButton(label string, onClick func()) *Buffer {
	toks  := make([]string, 1)
	lines := make([]span, 1)
	toks[0] = label
	b := &Buffer{
		bufferType: buttonT,
		Toks: toks,
		Lines: lines,
		X: nextButtonPos.X,
		Y: nextButtonPos.Y,
		W: len(label),
		H: 1,
		ID: newID,
		BorderStyle: stButtonBorder,
	}
	b.Draw		= b.buttonDraw
	b.Write		= b.buttonWrite
	b.OnHot		= b.buttonHot
	b.OnClick	= onClick
	b.OnUpdate	= func(){}
	b.box()

	buffers = append(buffers, b)
	newID++
	nextButtonPos.X = b.X + b.W + 2 

	return b
}

func NewWatcher(label string, v any) *Buffer {
	toks      := make([]string, 1)
	lines     := make([]span, 1)
	b := &Buffer{
		Name: label,
		bufferType: watcherT,
		Toks: toks,
		Lines: lines,
		X: nextWatcherPos.X,
		Y: nextWatcherPos.Y,
		H: 3,
		ID: newID,
		BorderStyle: stWatcherBorder,
	}
	b.Watch.current = v
	b.Draw  = b.textDraw
	b.Write = b.textWrite
	b.OnHot = b.textOnHot
	b.OnUpdate = func(){
		v := reflect.ValueOf(b.Watch.current)
		if v.Elem().Interface() != b.Watch.previous {
			state := fmt.Sprintf("%v\n", v.Elem().Interface())
			b.Write(state)
			b.Watch.previous = v.Elem().Interface()
			b.box()
			b.Draw()
		}
	}
	b.OnUpdate()
	
	last := len(b.Toks)-1
	b.W   = len(b.Name) + len(b.Toks[last])
	b.OnClick = func(){}
	buffers = append(buffers, b)
	b.box()
	newID++
	nextWatcherPos.Y = b.Y + b.H + 2
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
		// TODO: Deal with it later
		if s[i] == ' ' ||
		   s[i] == '\n' {
			tok := s[start: i+1]
			toks = append(toks, tok)
			start = i+1
		}
	}
	return toks
}

func mousePosChanged() bool {
	xChange := mouse.x != mouse.prev.x
	yChange := mouse.y != mouse.prev.y
	return xChange || yChange
}
