package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"

	//"runtime/pprof"
	"time"

	//"math/rand"

	"github.com/ilenker/fui"
)

// Template user side api testing
// with log dumps.

type Foo struct {
	Bar string
	Baz int
	SubFoo *Foo
}

func main() {
	//fCPUProfile, _ := os.Create("cpu.prof")

	Gang :=
	[]RPGCharacter{
		{
			Name:       "Frodo",
			Level:      1,
			Attributes: Stats{Strength: 10, Intelligence: 8, Dexterity: 6},
			Gear: Equipment{
				RightHand: &Item{Name: "Sword", Weight: 5, Value: 100},
				LeftHand:  &Item{Name: "Shield", Weight: 5, Value: 50},
				Armor:     Item{Name: "Leather Armor", Weight: 50, Value: 300},
			},
			Inventory: []Item{
				{Name: "Potion of Healing", Weight: 1, Value: 20},
				{Name: "Potion of Fireball", Weight: 1, Value: 30},
				{Name: "Potion of Invisibility", Weight: 1, Value: 40},
				{Name: "Potion of Slowness", Weight: 1, Value: 50},
			},
			Party: []*RPGCharacter{
				{
					Name:       "Sam",
					Level:      1,
					Attributes: Stats{Strength: 12, Intelligence: 10, Dexterity: 8},
					Gear: Equipment{
						RightHand: &Item{Name: "Dagger", Weight: 1, Value: 5},
						LeftHand:  &Item{Name: "Dagger", Weight: 1, Value: 5},
						Armor:     Item{Name: "Chain Mail", Weight: 50, Value: 200},
					},
					Inventory: []Item{
						{Name: "Potion of Healing", Weight: 1, Value: 20},
						{Name: "Potion of Fireball", Weight: 1, Value: 30},
						{Name: "Potion of Invisibility", Weight: 1, Value: 40},
						{Name: "Potion of Slowness", Weight: 1, Value: 50},
					},
				},
			},
		},
		{
			Name:       "Merry",
			Level:      1,
			Attributes: Stats{Strength: 14, Intelligence: 12, Dexterity: 10},
			Gear: Equipment{
				RightHand: &Item{Name: "Sword", Weight: 5, Value: 100},
				LeftHand:  &Item{Name: "Shield", Weight: 5, Value: 50},
				Armor:     Item{Name: "Leather Armor", Weight: 50, Value: 300},
			},
			Inventory: []Item{
				{Name: "Potion of Healing", Weight: 1, Value: 20},
				{Name: "Potion of Fireball", Weight: 1, Value: 30},
				{Name: "Potion of Invisibility", Weight: 1, Value: 40},
				{Name: "Potion of Slowness", Weight: 1, Value: 50},
			},
			Party: []*RPGCharacter{
				{
					Name:       "Pippin",
					Level:      1,
					Attributes: Stats{Strength: 16, Intelligence: 14, Dexterity: 12},
					Gear: Equipment{
						RightHand: &Item{Name: "Dagger", Weight: 1, Value: 5},
						LeftHand:  &Item{Name: "Dagger", Weight: 1, Value: 5},
						Armor:     Item{Name: "Chain Mail", Weight: 50, Value: 200},
					},
					Inventory: []Item{
						{Name: "Potion of Healing", Weight: 1, Value: 20},
						{Name: "Potion of Fireball", Weight: 1, Value: 30},
						{Name: "Potion of Invisibility", Weight: 1, Value: 40},
						{Name: "Potion of Slowness", Weight: 1, Value: 50},
					},
					Party: []*RPGCharacter{
						{
							Name:       "Gandalf",
							Level:      1,
							Attributes: Stats{Strength: 18, Intelligence: 16, Dexterity: 14},
							Gear: Equipment{
								RightHand: &Item{Name: "Sword", Weight: 5, Value: 100},
								LeftHand:  &Item{Name: "Shield", Weight: 5, Value: 50},
								Armor:     Item{Name: "Leather Armor", Weight: 50, Value: 300},
							},
							Inventory: []Item{
								{Name: "Potion of Healing", Weight: 1, Value: 20},
								{Name: "Potion of Fireball", Weight: 1, Value: 30},
								{Name: "Potion of Invisibility", Weight: 1, Value: 40},
								{Name: "Potion of Slowness", Weight: 1, Value: 50},
							},
						},
					},
				},
			},
		},
	}
	fui.Init()

	treeRoot, _ := fui.Tree("TreeView", Gang)
	fui.Button("level up", func(b *fui.Box) {
		Gang[0].Level++
		Gang[0].Party[0].Level++
		Gang[1].Level++
		Gang[1].Party[0].Level++
		Gang[1].Party[0].Party[0].Level++
	})
	//fui.Watcher("tree toks", tree.GetToks())

	//pprof.StartCPUProfile(fCPUProfile)
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

	nodesList := make([]*fui.TreeNode, 0)

	var foo func(n *fui.TreeNode)
	foo = func(n *fui.TreeNode){
		for _, c := range n.Children {
			if len(c.Children) != 0 {
				nodesList = append(nodesList, c)
				foo(c)
			}
		}
	}

	foo(treeRoot.Root)

	for _, n := range nodesList {
		n.Folded = true
	}

	t := time.NewTicker(time.Millisecond * 1000)
	for {
		select {
		case <-fui.ExitSig:
			//pprof.StopCPUProfile()
			return
		case <-t.C:
		//nodesList[rand.Intn(len(nodesList))].Folded =
		//!nodesList[rand.Intn(len(nodesList))].Folded
		//	fui.Redraw = true
		}
	}
}

// Structs for testing -------------------------------------------------------------------------------------
// System example ----------------------------------------------------------------------------------------------------
type SystemStatus struct {
	HostName      string
	UptimeSeconds uint64
	BatteryLevel  float32
	IsCharging    bool
	CoreTemp      int
	ProcessID     int
}

// Project example ---------------------------------------------------------------------------------------------------
type User struct {
	Username string
	Email    string
}

type Task struct {
	ID         int
	Title      string
	IsComplete bool
	Tags       []string
	AssignedTo User
}

type ProjectBoard struct {
	ProjectName string
	SprintID    int
	Backlog     []Task
}

// RPG example ------------------------------------------------------------------------------------------------------
type Stats struct {
	Strength     int
	Intelligence int
	Dexterity    int
}

type Item struct {
	Name   string
	Weight float64
	Value  int
}

type Equipment struct {
	RightHand *Item
	LeftHand  *Item
	Armor     Item
}

type RPGCharacter struct {
	Name       string
	Level      int
	Attributes Stats
	Gear       Equipment
	Inventory  []Item
	Party      []*RPGCharacter
}

func explode(v reflect.Value, maxDepth int) {
	var walk func(reflect.Value, string, int)
	walk = func(v reflect.Value, label string, d int) {
		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			if v.IsNil() {
				return
			}
		}
		if !v.IsValid() {
			return
		}
		if d >= maxDepth {
			return
		}
		switch v.Kind() {
		case reflect.Map:
			//panic("Map not supported")
		case reflect.Chan:
			//panic("Channel not supported")
		case reflect.UnsafePointer:
			//panic("Unsafe Pointer not supported")
		case reflect.Array:
			return
			l := v.Len()
			for i := range l {
				walk(v.Index(i), "", d+1)
			}
		case reflect.Slice:
			if v.Index(0).Kind() == reflect.Struct || v.Index(0).Kind() == reflect.Pointer {
				l := v.Len()
				for i := range l {
					walk(v.Index(i), "", d+1)
				}
			} else {
				w := fui.Watcher(label, &v)
				w.Y += d * 4
			}
		case reflect.Struct:
			fs := reflect.VisibleFields(v.Type())
			for _, f := range fs {
				walk(v.FieldByIndex(f.Index), f.Name, d+1)
			}
		case reflect.Pointer:
			v := v.Elem()
			walk(v, label, d)
		default:
			fui.Watcher(label, &v)
		}
	}
	walk(v, "", 0)
}
