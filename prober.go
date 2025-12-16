package prober

import (
	"time"
	"github.com/ilenker/prober/internal/calc"
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
	x, y int
	mask tc.ButtonMask
	prev struct{
		x, y int
		mask tc.ButtonMask
	}
}

//   - Increments whenever a buffer is made.
//   - Provides unique id that corresponds
//   - to it's index in the slice of buffers.
//~~~
var newID int
var scrW  int
var scrH  int

var nextButtonPos   calc.Vec2
var nextTerminalPos calc.Vec2
var nextWatcherPos  calc.Vec2
// Global states
var (
	redraw  bool
	exit    bool
	active  bool
)

func Init() {
	setCOLORTERM()
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()
	active = true
	initStyles()

	scrW, scrH = scr.Size()
	nextButtonPos.X   = 3
	nextButtonPos.Y   = scrH - 5
	nextTerminalPos.X = 3
	nextTerminalPos.Y = 2
	nextWatcherPos.X  = scrW - 20
	nextWatcherPos.Y  = 2

	buffers = make([]*Buffer, 0, 8)
	tick = time.NewTicker(hz)
	RedrawAll()
}


func Start() {
	// Printf debugs on exit
	defer scr.Fini()
	defer restoreCOLORTERM()
	go func() {
		for {
			ev := <-stdin
			if time.Since(ev.When()) > time.Millisecond*10 {
				continue
			}
			if mev, ok := ev.(*tc.EventMouse); ok {
				mouse.x, mouse.y = mev.Position()
				mouse.mask = mev.Buttons()
				for _, b := range buffers {
					b.OnHot(mev)
				}
				mouse.prev.x, mouse.prev.y = mev.Position()
				mouse.prev.mask = mev.Buttons()
			}
			if key, ok := ev.(*tc.EventKey); ok {
				if key.Key() == tc.KeyEsc {
					exit = true
					active = false
					return
				}
				if focusedIdx == -1 {
					continue
				}
			}
		}
	}()

	for !exit {
		if !active {
			time.Sleep(time.Millisecond * 50)
		}
		<-tick.C
		if redraw {
			scr.Clear()
			RedrawAll()
		}
		OnUpdateAll()
		scr.Show()
	}

}

func Resume() {
	active = true
	scr.Resume()
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
