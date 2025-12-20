package main

import (
	"syscall"
	"fmt"
	"runtime/debug"
	"os"
	"github.com/ilenker/fui"
)

// Template user side api testing
// with log dumps.
func main() {
	/* ------------ logging ------------ */
	f, _ := os.Create("stderr_capture.log")
	syscall.Dup2(int(f.Fd()), 2)
	/* --------------------------------- */
	fui.Init()

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

	select{}
}
