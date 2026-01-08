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
	ui.Layout.Rects[tty2].X += 30
	ui.Layout.Rects[tty2].W += 20
	ui.Terms.WriteToBuffer(tty, string(tests["all"]))
	//ui.Terms.WriteToBuffer(tty, "Hello world!")

	// Start the UI in a goroutine
	go fui.Start()

	// Simulating a main application loop
	t := time.NewTicker(16 * time.Millisecond)
	s := time.NewTicker(2 * time.Second)
	for {
		// Check for UI exit signal
		select {
		case <-fui.ExitSig:
			return
		case <-t.C:
			s := fmt.Sprintf("hz:%s|az:%s|h:%d|a:%d\nview:%d",
				fui.Mouse.HotZone.String(),
				fui.Mouse.ActZone.String(),
				fui.Mouse.HotID,
				fui.Mouse.ActID,
				ui.Terms.Views[tty])
			ui.Terms.Buffers[tty2] = []byte(s)
		case <-s.C:
		}
	}
}

var tests = map[string][]byte{
	"all": []byte("Go: α -> 世界 -> 🚀"),
    "ascii_only": []byte("12345"), 
    
    // WIDTH CHECKS
    "wide_cjk":   []byte("世"),      // 3 bytes, Width: 2
    "wide_emoji": []byte("🍣"),      // 4 bytes, Width: 2
    "narrow_sym": []byte("€"),       // 3 bytes, Width: 1 (Euro sign)
    
    // COMBINING CHARACTERS (Crucial for cursor positioning)
    // 'n' (width 1) + combining tilde (width 0) = Visual 'ñ'
    // If your loop adds widths blindly, this might count as 1 or 2 depending on logic.
    "combining": []byte{'n', 0xCC, 0x83}, 

    // TABULATION
    // Tabs are 1 byte but variable width. 
    // runewidth.RuneWidth('\t') usually returns 0 or 1, but terminal renders it as N.
    "control": []byte("Col1\tCol2"),
}

var hobbit = `In a hole in the ground there lived a hobbit. Not a nasty, dirty, wet hole, filled with the ends of worms and an oozy smell, nor yet a dry, bare, sandy hole with nothing in it to sit down on or to eat: it was a hobbit-hole, and that means comfort.

It had a perfectly round door like a porthole, painted green, with a shiny yellow brass knob in the exact middle. The door opened on to a tube-shaped hall like a tunnel: a very comfortable tunnel without smoke, with panelled walls, and floors tiled and carpeted, provided with polished chairs, and lots and lots of pegs for hats and coats—the hobbit was fond of visitors. The tunnel wound on and on, going fairly but not quite straight into the side of the hill—The Hill, as all the people for many miles round called it—and many little round doors opened out of it, first on one side and then on another. No going upstairs for the hobbit: bedrooms, bathrooms, cellars, pantries (lots of these), wardrobes (he had whole rooms devoted to clothes), kitchens, dining-rooms, all were on the same floor, and indeed on the same passage. The best rooms were all on the left-hand side (going in), for these were the only ones to have windows, deep-set round windows looking over his garden, and meadows beyond, sloping down to the river.`
