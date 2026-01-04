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

// NativeMenu returns the native menu handle associated with the given element
// or its nearest ancestor that is a [NativeMenuElement].
// If element is nil or there is no such ancestor, nil is returned.
func NativeMenu(ctx *Context, element Element) native.Handle {
	for elem := element; elem != nil; elem = elem.Parent() {
		if nativeElem, ok := elem.(NativeMenuElement); ok {
			return nativeElem.NativeMenu()
		}
	}
	return nil
}

// NativeMenuItem returns the native menu item handle associated with the given element
// or its nearest ancestor that is a [NativeMenuItemElement].
// If element is nil or there is no such ancestor, nil is returned.
func NativeMenuItem(ctx *Context, element Element) native.Handle {
	for elem := element; elem != nil; elem = elem.Parent() {
		if nativeElem, ok := elem.(NativeMenuItemElement); ok {
			return nativeElem.NativeMenuItem()
		}
	}
	return nil
}
