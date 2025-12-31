package fui

import (
	"encoding/json"
	"os"
)

type layoutEntry struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	W    int    `json:"w"`
	H    int    `json:"h"`
}

type layout []layoutEntry

/*
func saveLayout() error {
	layout := make(layout, len(boxes))
	for i := range boxes {
		if boxes[i] == nil {
			continue
		}
		entry := layoutEntry{
			Name: boxes[i].Name,
			Type: int(boxes[i].boxType),
			X:    boxes[i].X,
			Y:    boxes[i].Y,
			W:    boxes[i].W,
			H:    boxes[i].H,
		}
		layout[i] = entry
	}
	data, err := json.Marshal(layout)
	if err != nil {
		return err
	}
	os.WriteFile("layout_autogen.json", data, 0644)
	return nil
}
*/

func loadLayout() (bool, error) {
	// TODO: Recall boxes created dynamically via Struct Tree box
	data, err := os.ReadFile("layout_autogen.json")
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(data, &restoredLayout)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
func applyRestoredLayout() {
	for i, buf := range boxes {
		for j, entry := range restoredLayout {
			if entry.Name == buf.Name {
				boxes[i].X = entry.X
				boxes[i].Y = entry.Y
				if entry.Type != int(buttonT) {
					boxes[i].W = entry.W
					boxes[i].H = entry.H
				}
				restoredLayout[j].Name = "^=_$" + entry.Name
				boxes[i].reflowLines()
				break
			}
		}
	}
}
*/
