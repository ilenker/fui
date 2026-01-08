package fui

import (
	"fmt"
	"time"

	tc "github.com/gdamore/tcell/v3"
	"github.com/ilenker/fui/internal/calc"
)

var (
	focusedIdx      int
	scr             tc.Screen
	scrW            int
	scrH            int
	stdin           chan tc.Event
	Hz              = time.Second / 60
	frameTick		*time.Ticker
	restoredLayout  layout
	nextButtonPos   calc.Vec2
	nextTerminalPos calc.Vec2
	nextFieldPos    calc.Vec2 // Unused
	nextPadPos      calc.Vec2 // Unused
	Mouse           struct {
		X, Y int
		PrevX, PrevY int
		Mask     tc.ButtonMask
		PrevMask tc.ButtonMask
		HotZone  Zone
		ActZone  Zone
		HotID    int
		ActID int
	}
	// This can be used to prevent your main function from returning
	//	<-fui.ExitSig // Blocks - signal is sent when fui.Exit is set to true.
	ExitSig chan interface{}
)

// Global state
var (
	Redraw       bool
	Exit         bool
	layoutLoadOK bool
)

var ctx *UI

func Init() *UI {
	// Tcell Setup
	var err error
	setCOLORTERM()
	scr, _ = tc.NewScreen()
	scr.Init()
	scr.EnableMouse()
	stdin = scr.EventQ()

	// Default layout settings
	scrW, scrH = scr.Size()
	nextButtonPos.X = 3
	nextButtonPos.Y = scrH - 5
	nextTerminalPos.X = 3
	nextTerminalPos.Y = 2
	layoutLoadOK, err = loadLayout()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	INIT_BOXES := 32

	ctx = &UI{
		Layout: LayoutData{
			Rects:      make([]Rect   , INIT_BOXES),
			RectsPrev:  make([]Rect   , INIT_BOXES),
			Types:      make([]BoxType, INIT_BOXES),
		},
		Terms: TerminalData{
			Buffers:  make([][]byte, INIT_BOXES),
			Lines:    make([][]Span, INIT_BOXES),
			Views:    make([]int   , INIT_BOXES),
			CursorXs: make([]int   , INIT_BOXES),
			CursorYs: make([]int   , INIT_BOXES),
		},
		Buttons: ButtonData{
			UserFunctions: make([]uint8, INIT_BOXES),
		},
		Names: make([]string, INIT_BOXES),
	}
	frameTick = time.NewTicker(Hz)
	ExitSig = make(chan any)
	return ctx
}

func Start() {
	defer func() {
		//err := saveLayout()
		//if err != nil {
		//	fmt.Printf("%v\n", err)
		//}
		scr.Fini()
		restoreCOLORTERM()
		select {
		case ExitSig <- nil:
		default:
		}
	}()

	if layoutLoadOK {
		//applyRestoredLayout()
	}
	redrawAll()

	go readInput()

	for !Exit {
		<-frameTick.C
		Redraw = true
		if Redraw {
			redrawAll()
		}
		scr.Show()
	}
}

func readInput() {
	for {
		ev := <-stdin
		if ev == nil {
			continue
		}
		if mev, ok := ev.(*tc.EventMouse); ok {
			Mouse.X, Mouse.Y = mev.Position()
			Mouse.Mask = mev.Buttons()

			// Find hot id
			ctx.UpdateMouseState()
			// We now know which box the mouse
			// is on, and which zone.
			// Now, on button, we check hot.
			ctx.ApplyMouseState()
			Mouse.PrevX, Mouse.PrevY = mev.Position()
			Mouse.PrevMask = mev.Buttons()
		}
		if key, ok := ev.(*tc.EventKey); ok {
			switch {
			case key.Key() == tc.KeyESC:
				Exit = true
				return

			case key.Key() == tc.KeyTAB:
			// TODO: Tab targeting

			case focusedIdx == -1:
				continue
			}
		}
	}
}

func onUpdateAll() {
}

func redrawAll() {
	if ctx.Count == 0 {
		Redraw = false
		return
	}
	scr.Clear()
	for id := range ctx.Count {
		switch ctx.Layout.Types[id] {
		case TerminalT:
			DrawTerminal(id, ctx)
		case ButtonT:
		}
	}
	Redraw = true
	if focusedIdx == -1 {
		return
	}
	// Draw focused box last
	//boxes[focusedIdx].border()
	//boxes[focusedIdx].Draw()
}

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}
