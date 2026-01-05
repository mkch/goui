package goui

import "github.com/mkch/goui/native"

// NativeMenuElement is an [Element] that represents a native menu.
type NativeMenuElement interface {
	Element
	// NativeMenu returns the native handle of the menu.
	NativeMenu() native.Handle
}

type NativeMenuItemElement interface {
	Element
	// NativeMenuItem returns the native handle of the menu item.
	NativeMenuItem() native.Handle
}
