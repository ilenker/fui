package fui

import (
	"fmt"
	"time"
	"runtime"

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
	slowTick		*time.Ticker
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

	debugString = make([]string, 32)
)

const INIT_BOXES = 32
const DEBUG_TTY = 31
var DEBUG = true

// Global state
var (
	Redraw       bool
	Exit         bool
	layoutLoadOK bool
)

var ctx *UI

func Init() *UI {
	// Tcell Setup
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
	//layoutLoadOK, err = loadLayout()
	//if err != nil {
	//	fmt.Printf("%v\n", err)
	//}

	ctx = &UI{
		Layout: LayoutData{
			Rects:      make([]Rect   , INIT_BOXES),
			RectsPrev:  make([]Rect   , INIT_BOXES),
			Types:      make([]BoxType, INIT_BOXES),
		},
		Terms: TerminalData{
			Buffers:       make([][]byte, INIT_BOXES),
			ShadowBuffers: make([][]byte, INIT_BOXES),
			Lines:         make([][]Span, INIT_BOXES),
			Views:         make([]int   , INIT_BOXES),
			CursorXs:      make([]int   , INIT_BOXES),
			CursorYs:      make([]int   , INIT_BOXES),
			LinkRanges:    make([]LinkRanges, INIT_BOXES),
		},
		Buttons: ButtonData{
			UserFunctions: make([]func(), INIT_BOXES),
		},
		Names: make([]string, INIT_BOXES),
	}

	if DEBUG {
		enableDebugTTY()
	}
	_ = fmt.Sprintf("")

	frameTick = time.NewTicker(Hz)
	slowTick  = time.NewTicker(Hz*8)
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
		select {
		case <-slowTick.C:
			debugString[0] = GetMemUsage()
		case <-frameTick.C:
			Redraw = true
			if Redraw {
				redrawAll()
			}
			debugDisplay()
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

func enableDebugTTY() {
	id := DEBUG_TTY
	ctx.Count++
	ctx.Layout.Rects[id] = Rect{ X: scrW/2, Y: 5, W: (scrW/2)-2, H: 35, }
	ctx.Layout.Types[id] = TerminalT
	ctx.Names[id] = "debug tty"
	ctx.Terms.Buffers[id] = make([]byte, 0, 1024)
	ctx.Terms.ShadowBuffers[id] = make([]byte, 0, 1024)
	ctx.Terms.Lines[id] = make([]Span, 0, 128)
	ctx.Terms.Views[id] = 0
	ctx.Terms.LinkRanges[id] = LinkRanges{
			Starts: make([]int, 0, 64),
			Ends:   make([]int, 0, 64),
			Data:   make([]string, 0, 64),
		}
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
			DrawButton(id, ctx)
		}
	}
	if DEBUG {
		DrawTerminal(DEBUG_TTY, ctx)
	}
	Redraw = true
	if focusedIdx == -1 {
		return
	}
	// Draw focused box last
	//boxes[focusedIdx].border()
	//boxes[focusedIdx].Draw()
}

func printf(v any, verb string) {
	f := "%" + verb + "\n"
	ctx.WriteToBuffer(DEBUG_TTY, fmt.Sprintf(f, v))
}

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func debugDisplay() {
	for i, s := range debugString {
		scr.PutStr(0, i, s)
	}
}

func GetMemUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	const toMB = 1024 * 1024
	return fmt.Sprintf("Alloc: %8.2f MB | Sys: %8.2f MB | GC: %4d",
		float64(m.Alloc)/float64(toMB),
		float64(m.Sys)/float64(toMB),
		m.NumGC,
	)
}
