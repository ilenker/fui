package main

import (
	"os/exec"
	"strings"
	"time"

	fui "github.com/ilenker/fui"
)

var fooInt = 0
var prompt *fui.Box

func main() {
	fui.Init()
	term := fui.Terminal("ttyvm")
	fui.Pad("hobbit", hobbittext)
	fui.Button("Greet",
		func(*fui.Box) {
			term.Write("Hello!\n")
			fooInt++
		})
	fui.Button("ls", func(*fui.Box) {
		s := subProc("ls -lh")
		term.Println(s)
	})
	prompt = fui.Prompt("$", func(*fui.Box) {
		l := prompt.Line(-1)
		term.Println(l)
		term.Println(subProc(l))
	})

	go fui.Start()

	fui.Watcher("foo", &fooInt)

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-fui.ExitSig:
			fooInt++
			return
		case <-t.C:
			fooInt++
		}
	}
}

func subProc(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	cmd := fields[0]
	var args []string
	if len(fields) > 1 {
		args = fields[1:]
	}
	ls := exec.Command(cmd, args...)
	b, _ := ls.Output()
	ls.Wait()
	return string(b)
}

var hobbittext = `In a hole in the ground there lived a hobbit. Not a nasty, dirty, wet hole, filled with the ends of worms and an oozy smell, nor yet a dry, bare, sandy hole with nothing in it to sit down on or to eat: it was a hobbit-hole, and that means comfort.

It had a perfectly round door like a porthole, painted green, with a shiny yellow brass knob in the exact middle. The door opened on to a tube-shaped hall like a tunnel: a very comfortable tunnel without smoke, with panelled walls, and floors tiled and carpeted, provided with polished chairs, and lots and lots of pegs for hats and coats—the hobbit was fond of visitors. The tunnel wound on and on, going fairly but not quite straight into the side of the hill—The Hill, as all the people for many miles round called it—and many little round doors opened out of it, first on one side and then on another. No going upstairs for the hobbit: bedrooms, bathrooms, cellars, pantries (lots of these), wardrobes (he had whole rooms devoted to clothes), kitchens, dining-rooms, all were on the same floor, and indeed on the same passage. The best rooms were all on the left-hand side (going in), for these were the only ones to have windows, deep-set round windows looking over his garden, and meadows beyond, sloping down to the river.`
