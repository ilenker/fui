package fui

import (
	"fmt"
	"math"
	"unicode/utf8"

	. "github.com/ilenker/fui/internal/calc"

	tc "github.com/gdamore/tcell/v3"
	rw "github.com/mattn/go-runewidth"
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
	ui.Terms.AttrRanges[id] = AttrRanges{
			Starts: make([]int, 0, 64),
			Ends:   make([]int, 0, 64),
			Attrs:  make([]mdFormatMask, 0, 64),
			Links:  make(map[int]string, 4),
		}
	return id
}

func (ui *UI) AddButton(name string, userFunction func()) int {
	id := ui.Count
	ui.Count++
	ui.Layout.Rects[id] = Rect{ X: 5, Y: 5, W: len(name), H: 1, }
	ui.Layout.RectsPrev[id] = Rect{ X: 5, Y: 5, W: len(name), H: 1, }
	ui.Layout.Types[id] = ButtonT
	ui.Names[id] = name
	ui.Buttons.UserFunctions[id] = userFunction
	return id
}

var boxDefault = [7]rune{'┌', '┐', '└', '┘', '│', '─', '○'}
var boxFocused = [7]rune{'┏', '┓', '┗', '┛', '┃', '━', '●'}

func (ui *UI) DrawBorder(id int) {
	rect := ui.Layout.Rects[id]
	boxType := ui.Layout.Types[id]
	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	var style tc.Style
	var rs *[7]rune

	switch id {
	case focusedIdx:
		style = stBorderFocused
		rs = &boxFocused
	default:
		switch boxType {
		case TerminalT:
			style = stTerminalBorder
		case ButtonT:
			style = stButtonBorder
		}
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
	Start    int
	End      int
	TabLevel int
}

type TerminalData struct {
	Buffers    [][]byte
	Lines      [][]Span
	Views      []int
	AttrRanges []AttrRanges
	CursorXs   []int
	CursorYs   []int
}

type AttrRanges struct {
	Starts	[]int
	Ends	[]int
	Attrs	[]mdFormatMask
	idx		int
	Links   map[int]string
}

func (td *TerminalData) AppendAttrRange(id, start, end int, attr mdFormatMask) {
	ars := &td.AttrRanges[id]
	ars.Starts = append(ars.Starts, start)
	ars.Ends   = append(ars.Ends  , end)
	ars.Attrs  = append(ars.Attrs , attr)
	ars.idx++
}

func (ui *UI) WriteToBuffer(id int, s string) {
	ui.Terms.Buffers[id] = append(ui.Terms.Buffers[id], []byte(s)...)
	ReflowLinesUTF8(&ui.Terms.Lines[id], ui.Terms.Buffers[id], ui.Layout.Rects[id].W)
}

func DrawTerminal(id int, ui *UI) {
	// Don't try rendering if the width is zero
	ui.DrawBorder(id)
	rect := ui.Layout.Rects[id]
	if rect.W <= 0 {
		return
	}
	if len(ui.Terms.Buffers[id]) == 0 {
		return
	}
	attr       := mdFormatMask(0)
	cursor     := 0
	attrCursor := 0
	y          := rect.Y
	yLimit     := rect.Y + rect.H - 1
	ts         := &ui.Terms
	buf        := ts.Buffers[id]
	meta       := &ts.AttrRanges[id]
	lineBreak  := false

	// For each line
	for i := ts.Views[id]; i < len(ts.Lines[id]); i++ {
		l := ts.Lines[id][i]
		idx := l.Start
		// draw indents
		for range l.TabLevel {
			for t := range TAB_SIZE {
				var tab rune
				tab = rune('0'+t)
				tab = ' '
				scr.SetContent(
					rect.X+cursor, y, tab, nil,
					stDef,
					)
				cursor++
			}
		}
		
		// Range Setup: find range that ends after byte index
		for attrCursor < len(meta.Starts) && meta.Ends[attrCursor] <= idx {
			attrCursor++
		}
		// Check if we are in a range
		if attrCursor < len(meta.Starts) && meta.Starts[attrCursor] <= idx {
			attr = meta.Attrs[attrCursor]
		} else {
			attr = 0
		}

		// For each rune in line
		for idx < l.End {
			for attrCursor < len(meta.Starts) {
				isInside := (attr != 0)
				if isInside {
					if idx == meta.Ends[attrCursor] {
						attr = 0
						attrCursor++ 
						continue 
					}
				} else {
					if idx == meta.Starts[attrCursor] {
						attr = meta.Attrs[attrCursor]
						continue 
					}
				}
				break
			}
			r, size := utf8.DecodeRune(buf[idx:])
			visualWidth := rw.RuneWidth(r)
			//attrClean, isLink := ProcessAttr(attr)
			st := AttrToStyle(attr)
			if r == '\n' {
				y++
				lineBreak = true
			} else {
				scr.SetContent(
					rect.X+cursor, y, r, nil,
					st,
					)
				lineBreak = false
			}
			idx += size
			cursor += visualWidth
		}
		cursor = 0
		if !lineBreak {
			y++
		}
		if y > yLimit {
			break
		}
	}
}

func DrawButton(id int, ui *UI) {
	// Don't try rendering if the width is zero
	ui.DrawBorder(id)
	rect := ui.Layout.Rects[id]
	if rect.W <= 0 {
		return
	}
	s := ui.Names[id]
	scr.PutStr(rect.X, rect.Y, s)
}

//func ReflowLines_(lines *[]Span, buf []byte, width int) {
//	// Clear lines for overwrite
//	*lines = (*lines)[:0]
//	lineStart := 0
//	prevSafeEnd := 0
//	for i := range buf {
//		if buf[i] == '\n' {
//			*lines = append(*lines, Span{lineStart, i})
//			lineStart = i+1
//			prevSafeEnd = lineStart
//			continue
//		}
//		currentLineWidth := 1+i-lineStart
//		if buf[i] == ' ' {
//			// Confirm if current line width is within limit
//			if currentLineWidth <= width {
//				prevSafeEnd = i
//			} else {
//				// Does not fit, use prevSafeEnd as end
//				*lines = append(*lines, Span{lineStart, prevSafeEnd})
//				lineStart = prevSafeEnd+1
//				prevSafeEnd++
//			}
//			continue
//		}
//		if currentLineWidth >= width {
//			if prevSafeEnd == lineStart {
//				prevSafeEnd = i
//			}
//			*lines = append(*lines, Span{lineStart, prevSafeEnd})
//			lineStart = prevSafeEnd+1
//			prevSafeEnd++
//		}
//	}
//	*lines = append(*lines, Span{lineStart, len(buf)})
//}

const TAB_SIZE = 4

func ReflowLinesUTF8(lines *[]Span, buf []byte, maxWidth int) {
	if maxWidth < 3 {
		return
	}
	// Clear line len for overwrite
	*lines = (*lines)[:0]
	runeIdx := 0
	tabLevel := 0
	lineStart := 0
	safeLineEnd := -1
	runningTotalVisualWidth := 0

	for runeIdx < len(buf) {
		var visualWidth int
		r, byteWidth := utf8.DecodeRune(buf[runeIdx:])
		visualWidth = rw.RuneWidth(r)
		totalTabSpace := tabLevel*TAB_SIZE
		maybeNewWidth := runningTotalVisualWidth + visualWidth + totalTabSpace

		if buf[runeIdx] == '\t' {
			if maybeNewWidth + TAB_SIZE + 1 < maxWidth {
				tabLevel++
			} else {
				tabLevel = 0
			}
			//visualWidth = TAB_SIZE
		} else {
		}

		if maybeNewWidth >= maxWidth {
			if safeLineEnd != -1 {
				*lines = append(*lines, Span{lineStart, safeLineEnd, tabLevel})
				lineStart = safeLineEnd + 1
				runeIdx = lineStart
			} else {
				// No break opportunities found yet
				*lines = append(*lines, Span{lineStart, runeIdx, tabLevel})
				lineStart = runeIdx
			}
			runningTotalVisualWidth = 0
			safeLineEnd = -1
			continue
		}
		// Check non-printables
		if r == '\n' {
			runeIdx += 1
			*lines = append(*lines, Span{lineStart, runeIdx, tabLevel})
			lineStart = runeIdx
			safeLineEnd = -1
			runningTotalVisualWidth = 0
			tabLevel = 0
			continue
		}
		//if r == '\t' {
		//	runeIdx++
		//	continue
		//}
		// Check for break character (space, dash, anything else specified)
		if r == ' ' /*|| r == '-'*/ {
			safeLineEnd = runeIdx
		}
		runningTotalVisualWidth += visualWidth
		runeIdx += byteWidth
	}
	// Clean up, add last line
	if lineStart < len(buf) {
		*lines = append(*lines, Span{lineStart, len(buf), tabLevel})
	}
}

// Not sure about this yet
type ButtonData struct {
	UserFunctions []func()
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
			switch zone {
			case ZoneInside:
				if ui.Layout.Types[i] == TerminalT {
					ui.FindAndHighlightRange(i, mx, my)
				}
			}
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
			focusedIdx = Mouse.ActID
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
				if ui.Layout.Types[Mouse.ActID] == ButtonT {
					return
				}
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
					ReflowLinesUTF8(&ui.Terms.Lines[Mouse.ActID], ui.Terms.Buffers[Mouse.ActID], rect.W)
					if prevLenLines == 0 {
						return
					}
					ratio := float64(prevView) / float64(prevLenLines)
					adjustedView := float64(len(ui.Terms.Lines[Mouse.ActID])) * ratio
					ui.Terms.Views[Mouse.ActID] = int(math.Round(adjustedView))
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
		if Mouse.PrevMask == tc.Button1 {
			if Mouse.HotID != -1 {
				if ui.Layout.Types[Mouse.HotID] == ButtonT {
					ui.Buttons.UserFunctions[Mouse.HotID]()
					Redraw = true
				}
			}
		}
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

var lastHiliteIdx int
// TODO(Johan): Make more efficient
func (ui *UI) FindAndHighlightRange(id, mX, mY int) {
	lines := ui.Terms.Lines[id]
	rect  := ui.Layout.Rects[id]
	view  := ui.Terms.Views[id]
	ars   := ui.Terms.AttrRanges[id]
	x := mX-rect.X
	y := mY-rect.Y
	lineIdx := y + view
	if lineIdx < 0 || lineIdx >= len(lines) {
		return
	}
	startByte := lines[lineIdx].Start
	endByte   := lines[lineIdx].End
	indentWidth  :=  lines[lineIdx].TabLevel*(TAB_SIZE)
	cursorCell := x-indentWidth
	cursorByte := ui.Terms.nthCellInLineToByteIndex(id, cursorCell, startByte, endByte)
	if cursorByte == -1 {
		panic("ahhh")
	}
	debugString[1] = fmt.Sprintf("cursorCell: [%5d]", cursorCell)
	debugString[2] = fmt.Sprintf("cursorByte: [%5d]", cursorByte)

	for i := 0; i < ars.idx; i++ {
		// Reset hilite
		ars.Attrs[i] &= ^HILITE
		switch {
		case cursorByte < ars.Starts[i]:
			continue
		case (i+1 < ars.idx) && cursorByte >= ars.Starts[i+1]:
			continue
		case cursorByte < ars.Ends[i]:
			for i < ars.idx &&
				ars.Attrs[i] & LINK != 0 {
				ars.Attrs[i] |= HILITE
				i++
				debugString[0] = fmt.Sprintf("link: [%50s]", ars.Links[i])
			}
		}
	}
}

func (td *TerminalData) nthCellInLineToByteIndex(id, nthCell, startByteIdx, endByteIdx int) int {
	cellIndex := 0
	if startByteIdx == endByteIdx {
		return 0
	}
	for i := startByteIdx; i < endByteIdx; {
		r, byteSize := utf8.DecodeRune(td.Buffers[id][i:])
		vw := rw.RuneWidth(r)
		cellIndex += vw
		if nthCell <  cellIndex &&
		   nthCell >= cellIndex-vw {
			return i
		}
		i += byteSize
	}
	return 0
}
