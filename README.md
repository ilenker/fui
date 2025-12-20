# FUI

A library for quickly creating dynamic views into running Go programs. Like `printf` debugging using TUI widgets.

## Example Usage

```go
package main

import (
    "time"
    "github.com/ilenker/fui"
)

func main() {
    fui.Init()

    // 1. Create a Terminal to log output
    term := fui.Terminal("Logs")
    term.Println("Hello world!")

    // 2. Create a variable watcher
    counter := 0
    fui.Watcher("Counter", &counter)

    // 3. Create a button
    fui.Button("Reset", func(*fui.Box) {
    	counter = 0
    	term.Println("Counter reset")
    })

    // Start the UI in a goroutine
    go fui.Start()

    // Simulating a main application loop
    t := time.NewTicker(100 * time.Millisecond)
    for {
    	// Check for UI exit signal
    	select {
    	case <-fui.ExitSig:
    		return
    	case <-t.C:
        	counter++
    	}
    }
}
```

## Features

    Terminal: For printing to

    Prompt: For text input

    Watcher: For monitoring variables

    Button: For triggering functions on-click

    Layout Persistence: Boxes save their position and size to layout_autogen.json upon exit, restoring them on the next run.

There are some more examples in the demos/ directory:

    calculator/
        Simple calculator, with some watchers.
    sandbox/
        Shell command execution and widget tests.
    advent-of-code-2025-day10/
        A visualizer for an AoC 2025 solution.
