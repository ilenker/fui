package main

import (
	"time"

	tc "github.com/gdamore/tcell/v3"
)

var scr tc.Screen 
var stDef = tc.StyleDefault
var	stdin chan tc.Event
var buffers []*Buffer
var tick *time.Ticker
var hz = time.Second / 120
var focusedIdx int
var mouse struct{
	X, Y int
	Buttons tc.ButtonMask
	Prev struct{
		X, Y int
		Buttons tc.ButtonMask
	}
	Moved bool

}

// Global states
var REDRAW  bool
var EXIT    bool

func init() {
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()

	buffers = make([]*Buffer, 0, 8)
	focusedIdx = -1

	// Test buffers
	id := -1
	for i := range 6 {
		id++
		x, y := 3 + i*15, 3
		w, h := 10, 5
		size := 2048
		buffer := newTextBuffer(x, y, w, h, size, id)
		buffers = append(buffers, buffer)
	}

	id++
	x, y := 3, 12
	w, h := 7, 1
	size := 0
	onClick := func() {
		buffers[1].Write("Extra Text\n")
	}
	buffers[1].Write(text1)
	buffers[2].Write(text2)
	buffers[3].Write(text3)
	buffers[4].Write(text4)
	buffers[5].Write(text5)
	buffer := newButtonBuffer(onClick, x, y, w, h, size, id)
	buffer.Write("Button")
	buffers = append(buffers, buffer)

	tick = time.NewTicker(hz)
	RedrawAll()
}


func main() {
	defer scr.Fini()
	go func() {
		for {
			ev := <-stdin
			if time.Since(ev.When()) > time.Millisecond*10 {
				continue
			}
			if mev, ok := ev.(*tc.EventMouse); ok {
				for _, b := range buffers {
					b.OnMouseEvent(mev)
				}
			}
			if key, ok := ev.(*tc.EventKey); ok {
				if key.Key() == tc.KeyEsc {
					EXIT = true
					break
				}
				if focusedIdx == -1 {
					continue
				}
			}
		}
	}()

	for !EXIT {
		<-tick.C
		if REDRAW {
			scr.Clear()
			RedrawAll()
		}
		scr.Show()
	}
}


func RedrawAll() {
	for _, b := range buffers {
		if focusedIdx != b.ID {
			b.Box()
			b.OnUpdate()
			b.Draw()
		}
	}
	REDRAW = false
	if focusedIdx == -1 {
		return
	}
	// Draw focused buffer last
	buffers[focusedIdx].Box()
	buffers[focusedIdx].OnUpdate()
	buffers[focusedIdx].Draw()
}

var text0 = "Hello World Yes\nHello World Yes\n"
var text1 = "Hello\nWorld\nYes\nHello\nWorld\nYes\n"
var text2 = "\nHello World Yes Hello World Yes\n"
var text3 = "Hello  World  Yes\nHello World Yes\n"
var text4 = " Hello World Yes\nHello World Yes\n"
var text5 = "Hello World Yes\n\nHello World Yes\n"

var hobbittext =
`In a hole in the ground there lived a hobbit. Not a nasty, dirty, wet hole, filled with the ends of worms and an oozy smell, nor yet a dry, bare, sandy hole with nothing in it to sit down on or to eat: it was a hobbit-hole, and that means comfort.

It had a perfectly round door like a porthole, painted green, with a shiny yellow brass knob in the exact middle. The door opened on to a tube-shaped hall like a tunnel: a very comfortable tunnel without smoke, with panelled walls, and floors tiled and carpeted, provided with polished chairs, and lots and lots of pegs for hats and coats—the hobbit was fond of visitors. The tunnel wound on and on, going fairly but not quite straight into the side of the hill—The Hill, as all the people for many miles round called it—and many little round doors opened out of it, first on one side and then on another. No going upstairs for the hobbit: bedrooms, bathrooms, cellars, pantries (lots of these), wardrobes (he had whole rooms devoted to clothes), kitchens, dining-rooms, all were on the same floor, and indeed on the same passage. The best rooms were all on the left-hand side (going in), for these were the only ones to have windows, deep-set round windows looking over his garden, and meadows beyond, sloping down to the river.`
