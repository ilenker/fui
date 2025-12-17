package fui

import (
	"fmt"
	"time"
	"github.com/ilenker/fui/internal/calc"
	tc "github.com/gdamore/tcell/v3"
)

var scr tc.Screen 
var stDef = tc.StyleDefault
var	stdin chan tc.Event
var boxes []*Box
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
var restoredLayout layout

//   - Increments whenever a box is made.
//   - Provides unique id that corresponds
//   - to it's index in the slice of boxes.
//~~~
var newID int
var scrW  int
var scrH  int

var nextButtonPos   calc.Vec2
var nextTerminalPos calc.Vec2
var nextWatcherPos  calc.Vec2
// Global states
var (
	redraw         bool
	exit           bool
	active         bool
	restoreSuccess bool
)

func Init() {
	var err error
	setCOLORTERM()
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()
	active = true
	initStyles()

	// Default layout settings
	scrW, scrH = scr.Size()
	nextButtonPos.X   = 3
	nextButtonPos.Y   = scrH - 5
	nextTerminalPos.X = 3
	nextTerminalPos.Y = 2
	nextWatcherPos.X  = scrW - 20
	nextWatcherPos.Y  = 2
	restoreSuccess, err = loadLayout()
	if err != nil {
		fmt.Printf("%w\n", err)
	}

	boxes = make([]*Box, 0, 8)
	tick = time.NewTicker(hz)
}


func Start() {
	// Printf debugs on exit
	defer func() {
		err := saveLayout()
		if err != nil {
			fmt.Printf("%w\n", err)
		}
	}()
	defer scr.Fini()
	defer restoreCOLORTERM()

	if restoreSuccess {
		applyRestoredLayout()
	}
	scr.Clear()
	RedrawAll()

	go func() {
		for {
			ev := <-stdin
			if time.Since(ev.When()) > time.Millisecond*10 {
				continue
			}
			if mev, ok := ev.(*tc.EventMouse); ok {
				mouse.x, mouse.y = mev.Position()
				mouse.mask = mev.Buttons()
				for _, b := range boxes {
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

func OnUpdateAll() {
	for _, b := range boxes {
		b.OnUpdate()
	}
}


func RedrawAll() {
	if len(boxes) == 0 {
		return
	}
	for _, b := range boxes {
		if focusedIdx != b.ID {
			b.border()
			b.OnUpdate()
			b.Draw()
		}
	}
	redraw = false
	if focusedIdx == -1 {
		return
	}
	// Draw focused box last
	boxes[focusedIdx].border()
	boxes[focusedIdx].OnUpdate()
	boxes[focusedIdx].Draw()
}

var text0 = "Hello World Yes\nHello World Yes\n"
var text1 = "Hello\nWorld\nYes\nHello\nWorld\nYes\n"
var text2 = "\nHello World Yes Hello World Yes\n"
var text3 = "Hello  World  Yes\nHello World Yes\n"
var text4 = " Hello World Yes\nHello World Yes\n"
var text5 = "Hello World Yes\n\nHello World Yes\n"
