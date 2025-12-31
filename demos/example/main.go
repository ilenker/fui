package main

import (
	"fmt"
	"time"
	"github.com/ilenker/fui"
)

func main() {
	ui := fui.Init()
	tty := ui.AddTerminal("Terminal")
	tty2 := ui.AddTerminal("Terminal2")
	ui.Layout.Rects[tty2].X += 20
	ui.Layout.Rects[tty2].W += 20
	ui.Terms.WriteToBuffer(tty, "Hello!")

	// Start the UI in a goroutine
	go fui.Start()

	// Simulating a main application loop
	t := time.NewTicker(16 * time.Millisecond)
	for {
		// Check for UI exit signal
		select {
		case <-fui.ExitSig:
			return
		case <-t.C:
			s := fmt.Sprintf("hz:%s|az:%s|h:%d|a:%d",
				fui.Mouse.HotZone.String(),
				fui.Mouse.ActZone.String(),
				fui.Mouse.HotID,
				fui.Mouse.ActID)
			ui.Terms.Buffers[tty2] = []byte(s)
		}
	}
}

