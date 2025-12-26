package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/ilenker/fui"
)

// Template user side api testing
// with log dumps.
func main() {
	x := RPGCharacter{
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
			&RPGCharacter{
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
			&RPGCharacter{
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
			},
			&RPGCharacter{
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
			},
			&RPGCharacter{
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
	}
	fmt.Println(reflect.TypeOf(x))

	fui.Init()

	explode(reflect.ValueOf(x), 10)
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

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-fui.ExitSig:
			return
		case <-t.C:
		}
	}
}

// Structs for testing -------------------------------------------------------------------------------------

// System example ----------------------------------------------------------------------------------------------------
type SystemStatus struct {
	HostName      string
	UptimeSeconds uint64
	BatteryLevel  float32 // Visualization: Progress Bar?
	IsCharging    bool    // Visualization: Icon/Checkbox
	CoreTemp      int     // Visualization: Color-coded text (Red if > 80)
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
	Tags       []string // Visualization: Comma-separated strings or badges
	AssignedTo User     // Visualization: Inline struct
}

type ProjectBoard struct {
	ProjectName string
	SprintID    int
	Backlog     []Task // Visualization: List/Table
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
	RightHand *Item // Visualization: Check for nil!
	LeftHand  *Item // Visualization: Check for nil!
	Armor     Item
}

type RPGCharacter struct {
	Name       string
	Level      int
	Attributes Stats           // Nested Level 1
	Gear       Equipment       // Nested Level 1 with Pointers
	Inventory  []Item          // Nested Level 1 (Slice)
	Party      []*RPGCharacter // Recursive / Nested Level 2 (Watch out for infinite loops if visualizing!)
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
