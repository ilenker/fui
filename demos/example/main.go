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
