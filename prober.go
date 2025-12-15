package prober

import (
	"time"

	tc "github.com/gdamore/tcell/v3"
)

var scr tc.Screen 
var stDef = tc.StyleDefault
var	stdin chan tc.Event
var buffers []*Buffer
var tick *time.Ticker
var hz = time.Second / 60
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
var id int

// Global states
var redraw  bool
var exit    bool

func Init() {
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()

	buffers = make([]*Buffer, 0, 8)
	tick = time.NewTicker(hz)
	RedrawAll()
}


func Start() {
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
					exit = true
					break
				}
				if focusedIdx == -1 {
					continue
				}
			}
		}
	}()

	for !exit {
		<-tick.C
		if redraw {
			scr.Clear()
			RedrawAll()
		}
		OnUpdateAll()
		scr.Show()
	}
}

func OnUpdateAll() {
	for _, b := range buffers {
		b.OnUpdate()
	}
}


func RedrawAll() {
	if len(buffers) == 0 {
		return
	}
	for _, b := range buffers {
		if focusedIdx != b.ID {
			b.box()
			b.OnUpdate()
			b.Draw()
		}
	}
	redraw = false
	if focusedIdx == -1 {
		return
	}
	// Draw focused buffer last
	buffers[focusedIdx].box()
	buffers[focusedIdx].OnUpdate()
	buffers[focusedIdx].Draw()
}

var text0 = "Hello World Yes\nHello World Yes\n"
var text1 = "Hello\nWorld\nYes\nHello\nWorld\nYes\n"
var text2 = "\nHello World Yes Hello World Yes\n"
var text3 = "Hello  World  Yes\nHello World Yes\n"
var text4 = " Hello World Yes\nHello World Yes\n"
var text5 = "Hello World Yes\n\nHello World Yes\n"

