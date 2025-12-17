package main

import (
	fui "github.com/ilenker/fui"
)

var jolts []int
var sets *fui.Box
var button1 *fui.Box

func main() {
	jolts = []int{3, 5, 4, 7}
	fui.Init()
	sets = fui.NewTerminal("Sets")
	fui.NewWatcher("Jolts", &jolts)
	fui.NewWatcher("Sets Info", &sets)

	fui.NewButton("[___1]", func(b *fui.Box) {
		doButtonSet([]int{0,0,0,1})
		sets.Println(b.Name)
	})
	fui.NewButton("[_1_1]", func(b *fui.Box){
		doButtonSet([]int{0,1,0,1})
		sets.Println(b.Name)
	})
	fui.NewButton("[__1_]", func(b *fui.Box){
		doButtonSet([]int{0,0,1,0})
		sets.Println(b.Name)
	})
	fui.NewButton("[__11]", func(b *fui.Box){
		doButtonSet([]int{0,0,1,1})
		sets.Println(b.Name)
	})
	fui.NewButton("[1_1_]", func(b *fui.Box){
		doButtonSet([]int{1,0,1,0})
		sets.Println(b.Name)
	})
	fui.NewButton("[11__]", func(b *fui.Box){
		doButtonSet([]int{1,1,0,0})
		sets.Println(b.Name)
	})
	fui.NewButton("Half",  func(b *fui.Box){ halfAll() })
	fui.NewButton("Reset", func(b *fui.Box){ jolts = []int{3, 5, 4, 7} })

	go fui.Start()

	select {
	}
}

// (3) (1,3) (2) (2,3) (0,2) (0,1)

func doButtonSet(set []int) {
	for i := range set {
		jolts[i] -= set[i]
	}
}

func halfAll() {
	for i := range jolts {
		jolts[i] = jolts[i] / 2
	}
}
