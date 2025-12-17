package main

import (
	"fmt"
	"strconv"
	"runtime"
	"time"
	"unicode"
	"github.com/ilenker/fui"
)

var term *fui.Box
var m runtime.MemStats
var alloc uint64

func main() {
	fmt.Println("Hello world!")
	fui.Init()

	term = fui.NewTerminal("Term")

	for i := range 10 {
		name := string(byte(i+'0'))
		fui.NewButton(name, func(b *fui.Box) {
			term.Write(name)
		})
	}
	fui.NewButton("+", func(b *fui.Box) {
		term.Write(b.Name)
	})
	fui.NewButton("-", func(b *fui.Box) {
		term.Write(b.Name)
	})
	fui.NewButton("eval", func(b *fui.Box) {
		parseLine()
	})
	fui.NewWatcher("Mem(KB)", &alloc)

	go fui.Start()

	for{
		runtime.ReadMemStats(&m)
		alloc = m.Alloc / 1024
		time.Sleep(time.Millisecond * 500)
	}
}

func parseLine() {
	line := term.Line(-1)

	var lhs, rhs int
	var op rune

	for i, r := range line {
		if !unicode.IsNumber(r) {
			lhs, _ = strconv.Atoi(line[:i])
			rhs, _ = strconv.Atoi(line[i+1:])
			op = r
			break
		}
	}
	switch op {
	case '+':
		result := lhs + rhs
		term.Write(fmt.Sprintf("\n  Eval: %d\n", result))
	case '-':
		result := lhs - rhs
		term.Write(fmt.Sprintf("\n  Eval: %d\n", result))
	}

}
