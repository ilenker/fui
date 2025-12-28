package fui

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	. "github.com/ilenker/fui/internal/calc"

	tc "github.com/gdamore/tcell/v3"
	rw "github.com/mattn/go-runewidth"
)

type boxType int

const (
	terminalT boxType = iota
	buttonT
	watcherT
	promptT
	padT
	treeT
)

type span struct {
	start int
	end   int
}

type Box struct {
	toks    []string
	lines   []span
	history []int
	boxType boxType
	Name    string
	view    int
	X, Y    int
	W, H    int
	cursor  struct {
		X, Y int
	}
	id   int
	prev struct {
		X, Y int
		W, H int
	}
	BorderStyle tc.Style
	LabelStyle  tc.Style
	BodyStyle   tc.Style
	Write       func(s string)
	Draw        func()
	OnHot       func(*tc.EventMouse)
	OnClick     func(*Box)
	OnUpdate    func()
	OnCR        func(*Box)
	Watch       struct {
		current  any
		previous string
	}
	volatile    bool
}

func (b *Box) restoreLayout() {
	if !layoutLoadOK {
		return
	}
	for j, entry := range restoredLayout {
		if entry.Name == b.Name {
			b.X = entry.X
			b.Y = entry.Y
			if entry.Type != int(buttonT) {
				b.W = entry.W
				b.H = entry.H
			}
			restoredLayout[j].Name = "^=_$" + entry.Name
			b.reflowLines()
			break
		}
	}
}

// _Write -------------------------------------------------------------------------------------
func (b *Box) textWrite(s string) {
	toks := tokenize(s)
	b.toks = append(b.toks, toks...)

	// Just reflow everything everytime for now
	b.reflowLines()
	// Stick view to bottom only
	// if already at the bottom
	if b.view == len(b.lines)-b.H-1 {
		b.view = len(b.lines) - b.H
	}
	b.textDraw()
}

func (b *Box) fieldWrite(s string) {
	// This will always be one character
	// until we implement copy paste
	if len(s) == 0 {
		return
	}
	toks := tokenize(s)
	var lastTokI int
	var lastCharI int

	// If there are no tokens, add new and move on, no checks
	if len(b.toks) == 0 {
		b.toks = append(b.toks, toks...)
		goto skipCheck
	}

	lastTokI = len(b.toks) - 1
	lastCharI = len(b.toks[lastTokI]) - 1
	// Last token was terminated - just append new tokens
	if b.toks[lastTokI][lastCharI] == '\n' {
		b.toks = append(b.toks, toks...)
		// Last token unterminated - don't make new, add to it directly
	} else {
		b.toks[lastTokI] += toks[0]
	}

skipCheck:
	// Just reflow everything everytime for now
	b.reflowLines()
	if s == "\n" {
		b.view++
	} else {
		if b.view != len(b.lines) - 1 {
			b.view = len(b.lines) - 1
		}
	}
	b.textDraw()
}

func (b *Box) watcherWrite(s string) {
	//toks := tokenize(s)
	b.toks = append(b.toks, s)

	// Just reflow everything everytime for now
	// Stick view to bottom only
	// if already at the bottom
	if b.view == len(b.toks)-b.H-1 {
		b.view = len(b.toks)-b.H
	}
	b.watcherDraw()
}

func (b *Box) buttonWrite(s string) {
	toks := tokenize(s)
	b.toks = append(b.toks, toks...)
	b.buttonDraw()
}

// _Draw -------------------------------------------------------------------------------------

func (b *Box) textDraw() {
	// Don't try rendering if the width is zero
	if b.W <= 0 {
		return
	}
	if len(b.toks) == 0 {
		return
	}
	for i := b.view; i < len(b.lines); i++ {
		if i-b.view >= b.H {
			break
		}
		span := b.lines[i]
		builder := strings.Builder{}
		for t := span.start; t <= span.end; t++ {
			builder.WriteString(b.toks[t])
		}
		x := 0
		for _, r := range builder.String() {
			y := i - b.view
			if x >= b.W {
				break
			}
			scr.SetContent(b.X+x, b.Y+y, r, nil, b.BodyStyle)
			b.cursor.X = x
			b.cursor.Y = y
			if r > 255 {
				x += rw.RuneWidth(r)
			} else {
				x++
			}
		}
	}
}

func (b *Box) treeDraw() {
	// Don't try rendering if the width is zero
	if b.W <= 0 {
		return
	}
	//if len(b.toks) == 0 {
	//	return
	//}
	for i := b.view; i < len(b.toks); i++ {
		if i-b.view >= b.H {
			break
		}
		line := b.toks[i]
		scr.PutStr(b.X, b.Y+i-b.view, line)
	}
}

func (b *Box) watcherDraw() {
	// Don't try rendering if the width is zero
	if b.W <= 0 {
		return
	}
	//if len(b.toks) == 0 {
	//	return
	//}
	if Timestep {
		if len(b.history) == 0 {
			return
		}
		if len(timesteps) == 0 {
			return
		}
		for i := 0; i < len(b.history); i++ {
			if b.history[i] > timesteps[timestepView] {
				return
			}
			line := b.toks[i]
			scr.PutStr(b.X, b.Y+i, line)
		}
		return
	}
	for i := 0; i < len(b.toks); i++ {
		if i >= b.H {
			break
		}
		if i+b.view >= len(b.toks) {
			break
		}
		line := b.toks[i+b.view]
		scr.PutStr(b.X, b.Y+i, line)
	}
}

func (b *Box) blank() {
	limitX := b.X + b.W + 2
	limitY := b.Y + b.H + 2
	for x := b.X - 1; x < limitX; x++ {
		for y := b.Y - 1; y < limitY; y++ {
			scr.SetContent(x, y, ' ', nil, stDimmed)
		}
	}
}

// Border
var boxDefault = [7]rune{'┌', '┐', '└', '┘', '│', '─', '○'}
var boxFocused = [7]rune{'┏', '┓', '┗', '┛', '┃', '━', '●'}

func (b *Box) border() {
	x, y, w, h := b.X, b.Y, b.W, b.H
	var style tc.Style
	var rs *[7]rune

	switch b.id {
	case focusedIdx:
		style = b.BorderStyle.Foreground(tc.ColorRed)
		rs = &boxFocused
	default:
		style = b.BorderStyle
		rs = &boxDefault
	}
	// Sides
	var lenLines int
	if b.boxType == watcherT {
		lenLines = len(b.toks)
	} else {
		lenLines = len(b.lines)
	}
	f := ILerp(0, lenLines, float64(b.view))
	sliderPos := int(Lerp(0, h, f))
	for i := range h {
		scr.SetContent(x-1, y+i, rs[4], nil, style)
		scr.SetContent(x+w, y+i, rs[4], nil, style)
	}
	if b.boxType != buttonT {
		scr.SetContent(x+w, y+sliderPos, '█', nil, style)
	}
	// Top
	for i := range w {
		scr.SetContent(x+i, y-1, rs[5], nil, style)
	}
	// Bottom
	for i := range w {
		scr.SetContent(x+i, y+h, rs[5], nil, style)
	}
	// Corners
	switch b.boxType {
	case buttonT:
		scr.SetContent(x+w, y+h, rs[3], nil, style) // Bottom Right without handle
	default:
		scr.SetContent(x+w, y+h, rs[6], nil, style) // Bottom Right with handle
	}
	scr.SetContent(x-1, y-1, rs[6], nil, style) // Top Left with handle
	scr.SetContent(x+w, y-1, rs[1], nil, style) // Top Right
	scr.SetContent(x-1, y+h, rs[2], nil, style) // Bottom Left

	// Label
	switch b.boxType {
	// TODO: add indicator for when you aren't at the bottom of the scroll
	case watcherT:
		last := len(b.toks) - 1
		_y := y + h
		i := 0
		for _, r := range b.Name {
			if i >= b.W {
				scr.SetContent(x+i, _y, '…', nil, stDimmed)
				return
			}
			scr.SetContent(x+i, _y, r, nil, b.LabelStyle)
			i++
		}
		scr.SetContent(x+i, _y, '[', nil, stDimmed)
		i++
		for _, r := range b.toks[last] {
			if i >= b.W {
				scr.SetContent(x+i, _y, ']', nil, stDimmed)
				break
			}
			scr.SetContent(x+i, _y, r, nil, b.LabelStyle)
			i++
		}
		scr.SetContent(x+i, _y, ']', nil, stDimmed)
	case buttonT:
		// Do nothing
	case padT:
		for i, r := range b.Name {
			if i >= b.W {
				scr.SetContent(x+i, y-1, '…', nil, stDimmed)
				return
			}
			scr.SetContent(x+i, y-1, r, nil, stDimmed)
		}
	default:
		for i, r := range b.Name {
			if i >= b.W {
				scr.SetContent(x+i, y-1, '…', nil, stDimmed)
				return
			}
			scr.SetContent(x+i, y-1, r, nil, b.BodyStyle)
		}
	}
}

func (b *Box) clearBorder() {
	x, y, w, h := b.X, b.Y, b.W, b.H
	// Sides
	style := tc.StyleDefault
	for i := range h {
		scr.SetContent(x-1, y+i, ' ', nil, style)
		scr.SetContent(x+w, y+i, ' ', nil, style)
	}
	// Top/Bottom
	for i := range w {
		scr.SetContent(x+i, y-1, ' ', nil, style)
		scr.SetContent(x+i, y+h, ' ', nil, style)
	}
	// Corners
	scr.SetContent(x-1, y-1, ' ', nil, style)
	scr.SetContent(x+w, y-1, ' ', nil, style)
	scr.SetContent(x-1, y+h, ' ', nil, style)
	scr.SetContent(x+w, y+h, ' ', nil, style)
}

func (b *Box) buttonDraw() {
	// Don't try rendering if the width is zero
	if b.W <= 0 {
		return
	}
	x := 0
	for _, r := range b.Name {
		scr.SetContent(b.X+x, b.Y, r, nil, b.BodyStyle)
		x++
	}
}

// _Token & Line Processing --------------------------------------------------------------------------------

func tokenize(s string) []string {
	toks := make([]string, 0)
	// Scan over every character in the string
	start := 0
	for i := range len(s) {
		// If we reach end and it's not a newline
		// we may need to add to it later
		// perhaps a flag for "unterminated last token"
		if i == len(s)-1 {
			tok := s[start : i+1]
			toks = append(toks, tok)
			break
		}
		// Look for spaces or '\n'
		// Every excess space gets its own token right now.
		// TODO: Deal with it later
		if s[i] == ' ' ||
			s[i] == '\n' {
			tok := s[start : i+1]
			toks = append(toks, tok)
			start = i + 1
		}
	}
	return toks
}

func (b *Box) reflowLines() {
	if len(b.toks) == 0 {
		return
	}
	b.lines = b.lines[:0]
	chars := 0
	start := 0

	for i, tok := range b.toks {
		l := len(tok)
		if l == 0 {
			continue
		}
		if chars+l > b.W && i > start {
			newRange := span{start, i - 1}
			b.lines = append(b.lines, newRange)
			start = i
			chars = 0
		}
		chars += l
		if tok[l-1] == '\n' {
			newRange := span{start, i}
			b.lines = append(b.lines, newRange)
			start = i + 1
			chars = 0
		}
	}
	// Trailing line
	if start < len(b.toks) {
		newRange := span{start, len(b.toks) - 1}
		b.lines = append(b.lines, newRange)
	}
}

// _OnHot --------------------------------------------------------------------------------

func (b *Box) terminalOnHot(ev *tc.EventMouse) {
	update := func() {
		Redraw = true
		b.reflowLines()
	}
	x, y := mouse.x, mouse.y
	moved := mousePosChanged()
	mods := ev.Modifiers()
	var lenLines int
	if b.boxType == watcherT {
		lenLines = len(b.toks)
	} else {
		lenLines = len(b.lines)
	}
	switch mouse.mask {
	case tc.Button1:
		// Move
		switch {
		case b.onTopLeft(x, y, 1):
			if focusedIdx == -1 {
				focusedIdx = b.id
			} else if focusedIdx != b.id {
				return
			}
			if moved {
				b.X, b.Y = x+1, y+1
				update()
			}
		case b.onBottomRight(x, y, 1):
			if focusedIdx == -1 {
				focusedIdx = b.id
			} else if focusedIdx != b.id {
				return
			}
			update()
			hPrev := b.H
			b.W, b.H = x-b.X, y-b.Y
			// Stick to the bottom of the view logic
			if b.view > 0 &&
				lenLines-b.view > b.H {
				b.view -= b.H-hPrev
			}
		case b.onScrollRegion(x, y):
			b.updateSlider(x, y)
		case b.inMe(x, y):
			if mods == tc.ModCtrl {
				b.view = lenLines - b.H
				if b.view < 0 {
					b.view = 0
				}
			}
		case focusedIdx == b.id:
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
				b.view = min(b.view+10, lenLines-1)
			} else {
				b.view = min(b.view+1, lenLines-1)
			}
			update()
		}
	default:
	}
}

func (b *Box) treeOnHot(ev *tc.EventMouse) {
	update := func() {
		Redraw = true
	}
	x, y := mouse.x, mouse.y
	moved := mousePosChanged()
	mods := ev.Modifiers()
	switch mouse.mask {
	case tc.Button1:
		// Move
		switch {
		case b.onTopLeft(x, y, 1):
			if focusedIdx == -1 {
				focusedIdx = b.id
			} else if focusedIdx != b.id {
				return
			}
			if moved {
				b.X, b.Y = x+1, y+1
				update()
			}
		case b.onBottomRight(x, y, 1):
			if focusedIdx == -1 {
				focusedIdx = b.id
			} else if focusedIdx != b.id {
				return
			}
			update()
			hPrev := b.H
			b.W, b.H = x-b.X, y-b.Y
			// Stick to the bottom of the view logic
			if b.view > 0 &&
				len(b.lines)-b.view > b.H {
				b.view -= b.H - hPrev
			}
		case b.onScrollRegion(x, y):
			return
			b.updateSlider(x, y)
		case b.inMe(x, y):
			return
			if mods == tc.ModCtrl {
				b.view = len(b.lines) - b.H
				if b.view < 0 {
					b.view = 0
				}
			}
		case focusedIdx == b.id:
			focusedIdx = -1
		}
	// Scroll up and down
	case tc.WheelUp:
		if b.inMe(x, y) {
			if mods == tc.ModCtrl {
				b.view = max(b.view-10, 0)
			} else {
				b.view = max(b.view-1, 0)
			}
			update()
		}
	case tc.WheelDown:
		if b.inMe(x, y) {
			if mods == tc.ModCtrl {
				b.view = min(b.view+10, len(b.toks)-1)
			} else {
				b.view = min(b.view+1, len(b.toks)-1)
			}
			update()
		}
	default:
		if b.inMe(x, y) && 
		mouse.prev.mask == tc.Button1 {
			relativeY := y - b.Y + b.view
			n, ok := b.Watch.current.(*TreeRoot).NodeYs[relativeY]
			if !ok {
				return
			}
			if mods == tc.ModCtrl {
				w := Watcher(n.Name, &n.Value)
				w.X = x + 3
				w.Y = y + 3
				w.volatile = true
				return
			}
			n.Folded = !n.Folded
			b.treeWrite()
			update()
		}
	}
}

func (b *Box) buttonHot(ev *tc.EventMouse) {
	Redraw = true
	x := mouse.x
	y := mouse.y

	switch mouse.mask {
	case tc.Button1:
		if focusedIdx == b.id {
			focusedIdx = -1
		}
	case tc.Button2:
		if b.onTopLeft(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.id
			} else if focusedIdx != b.id {
				return
			}
			b.X, b.Y = x+1, y+1
		} else if focusedIdx == b.id {
			focusedIdx = -1
		}
	default:
		if b.inMe(x, y) {
			if mouse.prev.mask == tc.Button1 {
				b.Flash()
				b.OnClick(b)
			}
		}
	}
}

func (b *Box) textOnHot(ev *tc.EventMouse) {
	x, y := mouse.x, mouse.y
	moved := mousePosChanged()
	mods := ev.Modifiers()

	update := func() {
		b.reflowLines()
	}
	switch mouse.mask {
	case tc.Button1:
		// Move
		if b.onTopLeft(x, y, 1) {
			if focusedIdx == -1 {
				focusedIdx = b.id
			} else if focusedIdx != b.id {
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
				focusedIdx = b.id
			} else if focusedIdx != b.id {
				return
			}
			update()
			b.W, b.H = x-b.X, y-b.Y
		} else if b.onScrollRegion(x, y) {
			b.updateSlider(x, y)
		} else if focusedIdx == b.id {
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
				b.view = min(b.view+10, len(b.lines)-1)
			} else {
				b.view = min(b.view+1, len(b.lines)-1)
			}
			update()
		}
	default:
	}
}

func (b *Box) setPrevs() {
	b.prev.X, b.prev.Y = b.X, b.Y
	b.prev.W, b.prev.H = b.W, b.H
}

func mousePosChanged() bool {
	xChange := mouse.x != mouse.prev.x
	yChange := mouse.y != mouse.prev.y
	return xChange || yChange
}
func (b *Box) updateSlider(x, y int) {
	f := ILerp(b.Y, b.Y+b.H, float64(y))
	f = Clamp(f, 0, 1)
	// TODO: Redo
	//b.ViewIdx = int(Lerp(0, len(b.lines)-1, f))
}
func (b *Box) onScrollRegion(x, y int) bool {
	if x == b.X+b.W &&
		y >= b.Y && y < b.Y+b.H {
		return true
	}
	return false
}
func (b *Box) onTopLeft(x2, y2, dist int) bool {
	xDist := Abs(b.X - 1 - x2)
	yDist := Abs(b.Y - 1 - y2)
	return yDist <= dist && xDist <= dist
}
func (b *Box) onBottomRight(x2, y2, dist int) bool {
	xDist := Abs(b.X + b.W - x2)
	yDist := Abs(b.Y + b.H - y2)
	return yDist <= dist && xDist <= dist
}
func (b *Box) inMe(x, y int) bool {
	inX := x >= b.X-1 && x <= b.X+b.W
	inY := y >= b.Y-1 && y <= b.Y+b.H
	return inX && inY
}

// =============================== Box _Types ===============================

/*
Spawn a new, basic terminal box.

Returns a pointer to the new box for calling it's methods. Example:

	term := fui.Terminal("Ttyvm")
	term.Println("Hello world!")
	term.Clear()
*/
func Terminal(name string) *Box {
	toks := make([]string, 0, 64)
	lines := make([]span, 0, 64)
	b := &Box{
		Name:        name,
		boxType:     terminalT,
		toks:        toks,
		lines:       lines,
		X:           nextTerminalPos.X,
		Y:           nextTerminalPos.Y,
		W:           15,
		H:           15,
		id:          nextID,
		BorderStyle: stTerminalBorder,
		LabelStyle:  stTerminalLabel,
	}
	b.Draw = b.textDraw
	b.Write = b.textWrite
	b.OnHot = b.terminalOnHot
	b.OnUpdate = func() {}

	boxes = append(boxes, b)
	nextID++
	nextTerminalPos.X = b.X + b.W + 2
	b.restoreLayout()
	return b
}

// TODO: docs
func Pad(name string, text string) *Box {
	toks := make([]string, 0, 64)
	lines := make([]span, 0, 64)
	b := &Box{
		Name:        name,
		boxType:     padT,
		toks:        toks,
		lines:       lines,
		X:           nextTerminalPos.X,
		Y:           nextTerminalPos.Y,
		W:           15,
		H:           15,
		id:          nextID,
		BorderStyle: stTextBorder,
		LabelStyle:  stTextLabel,
	}
	b.Draw = b.textDraw
	b.Write = b.textWrite
	b.OnHot = b.textOnHot
	b.OnUpdate = func() {}
	b.Write(text)

	boxes = append(boxes, b)
	nextID++
	nextTerminalPos.X = b.X + b.W + 2
	b.restoreLayout()
	return b
}

/*
Spawn a watcher that prints whenever value of "vPointer" changes. For example:

	v := 0
	fui.Watcher("value of v", &v)
	xs := []int{5, 6, 7, 8}
	fui.Watcher("int slice", &xs)
*/
func Watcher(label string, vPointer any) *Box {
	toks := make([]string, 0, 64)
	lines := make([]span, 0, 64)
	history := make([]int, 0, 1024)
	b := &Box{
		Name:        label,
		boxType:     watcherT,
		toks:        toks,
		lines:       lines,
		history:	 history,
		X:           nextWatcherPos.X,
		Y:           nextWatcherPos.Y,
		H:           3,
		BorderStyle: stWatcherBorder,
		LabelStyle:  stWatcherLabel,
	}
	b.Watch.current = vPointer
	b.Draw = b.watcherDraw
	b.Write = b.watcherWrite
	b.OnHot = b.terminalOnHot
	b.OnClick = func(*Box) {}

	// TODO: clean up nil / pointer check
	if vPointer == nil {
		b.OnUpdate = func() {}
	} else if reflect.ValueOf(vPointer).Kind() != reflect.Pointer {
		b.OnUpdate = func() {}
	} else {
		b.OnUpdate = func() {
			// TODO: figure out how we might switch on types and change the formatting verbs accordingly
			format := typeFormatTable(b.Watch.current)
			v := fmt.Sprintf(format, reflect.ValueOf(b.Watch.current).Elem().Interface())
			changed := v != b.Watch.previous
			if changed {
				b.watcherWrite(v + "\n")
				b.Watch.previous = v
				b.FlashLabel()
				b.border()
				b.watcherDraw()
				b.history = append(b.history, frameID)
				prevID, ok := Last(timesteps)
				if ok {
					if prevID != frameID {
						timesteps = append(timesteps, frameID)
					}
				} else {
					timesteps = append(timesteps, frameID)
				}
			}
		}
		b.OnUpdate()
	}

	last := len(b.toks) - 1
	b.W = len(b.Name) + len(b.toks[last])
	if len(deletedBoxes) > 0 {
		last := len(deletedBoxes)-1
		id := deletedBoxes[last]
		boxes[id] = b
		b.id = id
		deletedBoxes = deletedBoxes[:last]
	} else {
		b.id = nextID
		boxes = append(boxes, b)
		nextID++
	}
	nextWatcherPos.Y = b.Y + b.H + 2
	b.restoreLayout()
	return b
}

/*
Spawn a button that executes a function "onClick". For example:

	x := 0
	fui.Button("+1", func(b *fui.Box) {
		x++
	})
*/
func Button(label string, onClick func(self *Box)) *Box {
	toks := make([]string, 1)
	lines := make([]span, 1)
	b := &Box{
		Name:        label,
		boxType:     buttonT,
		toks:        toks,
		lines:       lines,
		X:           nextButtonPos.X,
		Y:           nextButtonPos.Y,
		W:           rw.StringWidth(label),
		H:           1,
		id:          nextID,
		BorderStyle: stButtonBorder,
		LabelStyle:  stButtonLabel,
	}

	b.Draw = b.buttonDraw
	b.Write = b.buttonWrite
	b.OnHot = b.buttonHot
	b.OnClick = onClick
	b.OnUpdate = func() {}

	boxes = append(boxes, b)
	nextID++
	nextButtonPos.X = b.X + b.W + 2
	b.restoreLayout()
	return b
}

/*
Spawn a text field that executes a function "onCR" (enter / carriage return / newline). For example:

	prompt := fui.Field("$", func(b *fui.Box) {
		terminal.Println(prompt.Line(-1))
	})
*/
func Prompt(name string, onCR func(self *Box)) *Box {
	toks := make([]string, 64)
	lines := make([]span, 64)
	b := &Box{
		Name:        name,
		boxType:     promptT,
		toks:        toks,
		lines:       lines,
		X:           nextButtonPos.X,
		Y:           nextButtonPos.Y,
		W:           15,
		H:           1,
		id:          nextID,
		BorderStyle: stButtonBorder,
		LabelStyle:  stFieldLabel,
	}
	b.Draw = b.textDraw
	b.Write = b.fieldWrite
	b.OnHot = b.terminalOnHot
	b.OnUpdate = func() {}
	b.OnCR = onCR

	boxes = append(boxes, b)
	nextID++
	nextButtonPos.X = b.X + b.W + 2
	b.restoreLayout()
	return b
}

// TODO: change return back to just *Box
func Tree(label string, vPointer any) (*TreeRoot, *Box) {
	toks := make([]string, 64)
	lines := make([]span, 64)
	b := &Box{
		Name:        label,
		boxType:     treeT,
		toks:        toks,
		lines:       lines,
		X:           nextWatcherPos.X,
		Y:           nextWatcherPos.Y,
		H:           15,
		W:			 15,
		id:          nextID,
		BorderStyle: stWatcherBorder,
		LabelStyle:  stWatcherLabel,
	}
	b.Watch.current = vPointer
	b.Draw = b.treeDraw
	b.Write = func(string) {}
	b.OnHot = b.treeOnHot
	b.OnClick = func(*Box) {}
	b.OnUpdate = func(){}

	// Save the root node (*StructNode) in Watch.current
	// Build our tokens from there
	tree := BuildTreeNodes(vPointer, 100)
	b.Watch.current = &TreeRoot{
		Name:   label,
		NodeYs: make(map[int]*TreeNode),
		Root:   tree,
	}
	// Redraw check cases:
	// OnClick  - very likely
	// OnHot    - possibly
	// OnUpdate - no

	b.treeWrite()

	boxes = append(boxes, b)
	nextID++
	nextWatcherPos.Y = b.Y + b.H + 2
	b.restoreLayout()
	return b.Watch.current.(*TreeRoot), b
}

// populates all tokens - 1 token per line
func (b *Box) treeWrite() {
	if b.boxType != treeT {
		return
	}
	// TODO: non nuclear option?
	clear(b.Watch.current.(*TreeRoot).NodeYs)
	b.toks = make([]string, 0)
	y := b.Y
	var draw func(*TreeNode, int)
	draw = func(node *TreeNode, d int) {
		indent := strings.Repeat(" ", d*2)
		for _, n := range node.Children {
			line := indent
			drawChildren := false
			if len(n.Children) != 0 && !n.Folded {
				drawChildren = true
				line += "-" + n.Name
			} else {
				if len(n.Children) == 0 {
					line += " " + n.Name
				} else {
					line += "+" + n.Name
				}
			}
			b.toks = append(b.toks, line)
			//scr.PutStr(
			//	b.X+d*2, y,
			//	line)
			b.Watch.current.(*TreeRoot).NodeYs[y-b.Y] = n
			y++
			if drawChildren {
				draw(n, d+1)
			}
		}
	}
	draw(b.Watch.current.(*TreeRoot).Root, 0)
}

// ================================ Box _Methods ===============================

// temporary cringe accessor
func (b *Box) GetToks() *[]string {
	return &b.toks
}

/*
Make the box invert colors for 100ms
*/
func (b *Box) Flash() {
	b.BodyStyle = b.BodyStyle.Reverse(true)
	Redraw = true
	go func() {
		t := time.NewTimer(time.Millisecond * 100)
		<-t.C
		b.BodyStyle = b.BodyStyle.Reverse(false)
		Redraw = true
	}()
}

func (b *Box) FlashLabel() {
	b.LabelStyle = b.LabelStyle.Reverse(true)
	Redraw = true
	go func() {
		t := time.NewTimer(time.Millisecond * 50)
		<-t.C
		b.LabelStyle = b.LabelStyle.Reverse(false)
		Redraw = true
	}()
}

func (b *Box) Backspace() {
	if len(b.toks) == 0 {
		return
	}
	s := b.toks[len(b.toks)-1]
	if len(s) == 0 {
		return
	}
	runes := []rune(s)
	// Remove the last rune (independent of width)
	if len(runes) == 0 {
		return
	}
	scr.SetContent(b.X+b.cursor.X, b.Y+b.cursor.Y, ' ', nil, b.BodyStyle)
	b.cursor.X = max(0, b.cursor.X-1)
	runes = runes[:len(runes)-1]
	b.toks[len(b.toks)-1] = string(runes)
}

func (b *Box) Println(s string) {
	// TODO: Docs
	if s == "" {
		return
	}
	toks := tokenize(s)
	b.toks = append(b.toks, toks...)
	b.toks[len(b.toks)-1] += "\n"

	// Just reflow everything everytime for now
	b.reflowLines()
	b.textDraw()
}

/* Clear makes a new line and sets the view to the bottom. */
func (b *Box) Clear() {
	b.Write("\n\n")
	b.view = len(b.lines)
	for x := b.X; x < b.X+b.W; x++ {
		for y := b.Y; y < b.Y+b.H; y++ {
			scr.SetContent(x, y, ' ', nil, stDimmed)
		}
	}
}

/* Returns the i-th line from a box as a string. Pass in -1 to get the newest line. */
func (b *Box) Line(i int) string {
	// TODO: access arbitrary lines
	if len(b.lines) == 0 ||
		len(b.toks) == 0 {
		return ""
	}
	if i != -1 {
		return ""
	}
	line := ""
	span := b.lines[len(b.lines)-1]
	for i := span.start; i <= span.end; i++ {
		line += b.toks[i]
	}
	return line
}

func (b *Box) Reset() {
	b.view = 0
	clear(b.lines)
	clear(b.toks)
}
