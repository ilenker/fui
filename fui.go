package fui

import (
	"fmt"
	"time"

	tc "github.com/gdamore/tcell/v3"
	"github.com/ilenker/fui/internal/calc"
)

var (
	boxes           []*Box
	nextID          int
	focusedIdx      int
	scr             tc.Screen
	scrW            int
	scrH            int
	stdin           chan tc.Event
	frameTick       *time.Ticker
	Hz              = time.Second / 60
	restoredLayout  layout
	nextButtonPos   calc.Vec2
	nextTerminalPos calc.Vec2
	nextWatcherPos  calc.Vec2
	nextFieldPos    calc.Vec2 // Unused
	nextPadPos      calc.Vec2 // Unused
	mouse           struct {
		x, y int
		mask tc.ButtonMask
		prev struct {
			x, y int
			mask tc.ButtonMask
		}
	}
	// This can be used to prevent your main function from returning
	//	<-fui.ExitSig // Blocks - signal is sent when fui.Exit is set to true.
	ExitSig chan interface{}
)

// Global state
var (
	Redraw       bool
	Exit         bool
	active       bool // Unused
	layoutLoadOK bool
)

func Init() {
	// Tcell Setup
	var err error
	setCOLORTERM()
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()
	active = true

	// Default layout settings
	scrW, scrH = scr.Size()
	nextButtonPos.X = 3
	nextButtonPos.Y = scrH - 5
	nextTerminalPos.X = 3
	nextTerminalPos.Y = 2
	nextWatcherPos.X = scrW - 20
	nextWatcherPos.Y = 2
	layoutLoadOK, err = loadLayout()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	boxes = make([]*Box, 0, 8)
	frameTick = time.NewTicker(Hz)
	ExitSig = make(chan interface{})
}

func Start() {
	defer func() {
		err := saveLayout()
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		scr.Fini()
		restoreCOLORTERM()
		select {
		case ExitSig <- nil:
		default:
		}
	}()

	if layoutLoadOK {
		applyRestoredLayout()
	}
	redrawAll()

	go readInput()

	for !Exit {
		if !active {
			time.Sleep(time.Millisecond * 50)
		}
		<-frameTick.C
		if Redraw {
			redrawAll()
		}
		onUpdateAll()
		scr.Show()
	}
}

func readInput() {
	for {
		ev := <-stdin
		if ev == nil {
			continue
		}
		// Forgot why we needed this
		//if time.Since(ev.When()) > time.Millisecond*10 {
		//	continue
		//}
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
			switch {
			case key.Key() == tc.KeyESC:
				Exit = true
				active = false
				return

			case key.Key() == tc.KeyTAB:
				focusedIdx = calc.WrapInt(focusedIdx+1, len(boxes))
				Redraw = true

			case focusedIdx == -1:
				continue

			// Pass keystrokes onto any focused text input field
			case boxes[focusedIdx].boxType == fieldT:
				switch key.Key() {
				case tc.KeyEnter:
					boxes[focusedIdx].OnCR(boxes[focusedIdx])
					boxes[focusedIdx].Write("\n")
					boxes[focusedIdx].reflowLines()
				case tc.KeyBackspace:
					boxes[focusedIdx].Backspace()
				default:
					boxes[focusedIdx].Write(key.Str())
				}
			}
		}
	}
}

func onUpdateAll() {
	for _, b := range boxes {
		b.OnUpdate()
	}
}

func redrawAll() {
	if len(boxes) == 0 {
		Redraw = false
		return
	}
	scr.Clear()
	for _, b := range boxes {
		if focusedIdx != b.id {
			b.border()
			b.OnUpdate()
			b.Draw()
		}
	}
	Redraw = false
	if focusedIdx == -1 {
		return
	}
	// Draw focused box last
	boxes[focusedIdx].border()
	boxes[focusedIdx].OnUpdate()
	boxes[focusedIdx].Draw()
}
