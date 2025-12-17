# FUI

Library for quickly creating dynamic views into programs.
Essentially printf debugging using widgets.

Example:
```
    newTerm := fui.NewTerminal("Term")
    newTerm.Println("Hello world!")

    fui.NewButton("Greeter", func(b *fui.Box){
        newTerm.Println(b.Name + " says hello")
    })

    foo := 0
    fui.NewWatcher("Foo", &foo)
```

This will spawn a terminal display, a button that triggers some function, and a box that prints whenever the value of foo changes, with a history to scroll back in.
They spawn in at their default locations, moving as needed. They can then be moved around and resized while the program runs. When you rerun the program, the layout will be restored to how it was left upon exit.
