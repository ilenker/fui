package fui

import (
	"reflect"
)

type TreeNode struct {
	Name     string
	Value    any
	Children []*TreeNode
	Folded   bool
}

type TreeRoot struct {
	Name   string
	Root   *TreeNode
	NodeYs map[int]*TreeNode
}

func BuildTreeNodes(vPointer any, maxDepth int) *TreeNode {
	value := reflect.ValueOf(vPointer)

	var walk func(reflect.Value, string, int) *TreeNode
	walk = func(v reflect.Value, label string, d int) *TreeNode {
		node := &TreeNode{
			Name: label,
			Value: v,
			Folded: true,
		}
		switch v.Kind() {
		case reflect.Chan,
			reflect.Func,
			reflect.Interface,
			reflect.Map,
			reflect.Pointer,
			reflect.Slice:
			if v.IsNil() {
				return node
			}
		}
		if !v.IsValid() {
			return node
		}
		if d >= maxDepth {
			return node
		}
		switch v.Kind() {
		case reflect.Map:
			panic("Map not supported")
		case reflect.Chan:
			panic("Channel not supported")
		case reflect.UnsafePointer:
			panic("Unsafe Pointer not supported")
		case reflect.Array:
			panic("Array not supported")
		case reflect.Slice:
			for i := range v.Len() {
				c := walk(v.Index(i), string(i+'0'), d+1)
				node.Children = append(node.Children, c)
			}
			return node
			//panic("Slice not supported")
		case reflect.Struct:
			fs := reflect.VisibleFields(v.Type())
			for _, f := range fs {
				c := walk(v.FieldByIndex(f.Index), f.Name, d+1)
				node.Children = append(node.Children, c)
			}
			return node
		case reflect.Pointer:
			// A pointer to what?
			v := v.Elem()
			switch v.Kind() {
			case reflect.Struct:
				fs := reflect.VisibleFields(v.Type())
				for _, f := range fs {
					c := walk(v.FieldByIndex(f.Index), f.Name, d+1)
					node.Children = append(node.Children, c)
				}
			}
		default:
			return node
		}
		return node
	}
	return walk(value, value.Type().String(), 0)
}
