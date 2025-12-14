package main

import (
	"fmt"
	"time"
	"unicode"

	tc "github.com/gdamore/tcell/v3"
)

var scr tc.Screen 
var stDef = tc.StyleDefault
var	stdin chan tc.Event
var buffers []*Buffer
var tick *time.Ticker
var hz = time.Second / 60

// Global states
var REDRAW  bool
var EXIT    bool

func init() {
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()

	buffers = make([]*Buffer, 0, 8)
	x, y, w, h, size := 3, 3, 5, 5, 2048*4
	buffer0 := newBuffer(x, y, w, h, size)
	//buffer0.Write(text)

	//x, y, w, h, size = 20, 3, 10, 10, 2048*4
	//buffer1 := newBuffer(x, y, w, h, size)
	//buffer1.Write(text)
	buffers = append(buffers, buffer0)
	tick = time.NewTicker(hz)
	RedrawBuffers()
}

func main() {
	defer scr.Fini()

	go func() {
		_n_ := 0
		for {
			_n_++
			scr.PutStr(0, 0, fmt.Sprintf("%10d", _n_))
			ev := <-stdin
			if mev, ok := ev.(*tc.EventMouse); ok {
				for _, b := range buffers {
					b.HandleMouseEvent(mev)
				}
			}
			if key, ok := ev.(*tc.EventKey); ok {
				if key.Key() == tc.KeyEsc {
					EXIT = true
					break
				}
				buffers[0].Write(key.Str())
			}
		}
	}()

	_n_ := 0
	for !EXIT {
		_n_++
		scr.PutStr(0, 1, fmt.Sprintf("%10d", _n_))
		<-tick.C
		if REDRAW {
			RedrawBuffers()
			REDRAW = false
		}
		scr.Show()
	}
}

type Buffer struct {
	Data    []rune
	i 	    int // The head of the data
	X, Y    int
	W, H    int
	ViewIdx int // From where in the data to start showing the buffer
	Cursor  struct{X, Y int}
}

func (b *Buffer) Write(s string) {
	// When we hit the end of the view
	// we want to open up a line in the
	// buffer.
	if b.i >= b.W * b.H &&
		b.i % b.W == 0 {
		b.ViewIdx += b.W*1
	}
	for i, r := range s {
		b.Data[b.i + i] = r
		if b.i >= len(b.Data) {
			b.i = 0 - i
			b.ViewIdx = i
		}
	}
	b.i += len(s)
	b.Draw()
}


func (b *Buffer) Draw() {
	X, Y, W, H := b.X, b.Y, b.W, b.H
	index := b.ViewIdx

	for y := Y;
		y <  Y + H;
		y++ {

		for x := X;
			x <  X + W;
			x++ {
			EOWDist := b.DistanceToEOW(index)
			if EOWDist > W - (x-X) {
				x = X
				goto NewLine
			}
			if index >= len(b.Data) {
				break
			}
			if b.Data[index] == '\n' {
				if b.Data[index+1] == '\n' {
					x = X
					index++
					goto NewLine
				}
			}
			scr.SetContent(
				x, y,
				b.Data[index],
				nil,
				stDef)
			index++
		}
		NewLine:
	}
}

func (b *Buffer) DrawInfo() {

}

func (b *Buffer) HandleMouseEvent(ev *tc.EventMouse) {
	defer func() {
		REDRAW = true
	}()
	x, y := ev.Position()
	button := ev.Buttons()
	cornerX := b.X + b.W
	cornerY := b.Y + b.H

	switch button {
	case tc.Button1:
		// Resize
		if InArea(cornerX, cornerY, x, y, 1) {
			b.Clear()
			b.W = x - b.X
			b.H = y - b.Y
			return
		}
		// Move
		if InArea(b.X, b.Y, x, y, 1) {
			b.Clear()
			b.X = x
			b.Y = y
			return
		}
	case tc.WheelUp:
		if InArea(b.X, 1, x, 1, 1) {
			b.ViewIdx -= b.W
			if b.ViewIdx < 0 {
				b.ViewIdx = 0
			}
			return
		}
	case tc.WheelDown:
		if InArea(b.X, 1, x, 1, 1) {
			b.ViewIdx += b.W
		}
		return
	default:
	}
}


func (b *Buffer) DistanceToEOW(from int) int {
	i := 0
	for {
		r := b.Data[from + i]
		if unicode.IsSpace(r) {
			return i
		}
		if from + i >= len(b.Data)-1 {
			return -1
		}
		i++
	}
}

func (b *Buffer) Clear() {
	return
	X, Y, W, H := b.X-1, b.Y-1, b.W+2, b.H+2

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

func (b *Buffer) Box() {
	x, y, w, h := b.X, b.Y, b.W, b.H
	// Sides
	style := tc.StyleDefault
	for i := range h {
		scr.SetContent(x - 1, y + i, '│', nil, style)
		scr.SetContent(x+w  , y + i, '│', nil, style)
	}
	// Top/Bottom
	for i := range w {
		scr.SetContent(x + i, y - 1, '─', nil, style)
		scr.SetContent(x + i, y+h  , '─', nil, style)
	}
	// Corners
	scr.SetContent(x - 1, y - 1, '┌', nil, style)
	scr.SetContent(x+w  , y - 1, '┐', nil, style)
	scr.SetContent(x - 1, y+h  , '└', nil, style)
	scr.SetContent(x+w  , y+h  , '┘', nil, style)
}

func RedrawBuffers() {
	for _, b := range buffers {
		b.Box()
		b.Draw()
	}
}


func newBuffer(x, y, w, h, size int) *Buffer {
	if size < w*h {
		size = w*h
	}
	data := make([]rune, size)
	b := &Buffer{
		Data: data,
		X: x,
		Y: y,
		W: w,
		H: h,
	}
	b.Box()
	return b
}


func InArea(x1, y1, x2, y2, dist int) bool {
	xDist := max(x1, x2) - min(x1, x2)
	yDist := max(y1, y2) - min(y1, y2)
	return yDist <= dist && xDist <= dist
}

var text =
`In a hole in the ground there lived a hobbit. Not a nasty,
dirty, wet hole, filled with the ends of worms and an oozy
smell, nor yet a dry, bare, sandy hole with nothing in it to
sit down on or to eat: it was a hobbit-hole, and that means
comfort.

It had a perfectly round door like a porthole, painted
green, with a shiny yellow brass knob in the exact middle.
The door opened on to a tube-shaped hall like a tunnel: a
very comfortable tunnel without smoke, with panelled
walls, and floors tiled and carpeted, provided with
polished chairs, and lots and lots of pegs for hats and
coats—the hobbit was fond of visitors. The tunnel wound
on and on, going fairly but not quite straight into the side
of the hill—The Hill, as all the people for many miles
round called it—and many little round doors opened out
of it, first on one side and then on another. No going
upstairs for the hobbit: bedrooms, bathrooms, cellars,
pantries (lots of these), wardrobes (he had whole rooms
devoted to clothes), kitchens, dining-rooms, all were on
the same floor, and indeed on the same passage. The best
rooms were all on the left-hand side (going in), for these
were the only ones to have windows, deep-set round
windows looking over his garden, and meadows beyond,
sloping down to the river.`
