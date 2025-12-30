package main

import (
	"fmt"
	"github.com/ilenker/fui"
	"os"
	"runtime/debug"
	"syscall"
)

// Template user side api testing
// with log dumps.
func main() {
	/* ------------ logging ------------ */
	f, _ := os.Create("stderr_capture.log")
	syscall.Dup2(int(f.Fd()), 2)
	/* --------------------------------- */
	fui.Init()

	x, y, z := 0, 0, 0
	fui.Watcher("x", &x)
	fui.Watcher("y", &y)
	fui.Watcher("z", &z)

	fui.Button("x++", func(*fui.Box) { x++ })
	fui.Button("y++", func(*fui.Box) { y++ })
	fui.Button("z++", func(*fui.Box) { z++ })
	fui.Button("xyz++", func(*fui.Box) { x++; y++; z++ })

	/* ------------ logging ------------ */
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fui.Exit = true
				stack := debug.Stack()
				errLog := fmt.Sprintf("panic: %v\n\n%s", r, stack)
				_ = os.WriteFile("crash.log", []byte(errLog), 0644)
				os.Exit(1)
			}
		}()
		fui.Start()
	}()
	/* --------------------------------- */

	select {
	case <-fui.ExitSig:
	}
}
