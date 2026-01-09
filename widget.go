package goui

import "github.com/mkch/goui/marker"

// AbstractWidget is the base interface for all widgets.
// It lacks any exclusive method to identify its kind or type.
type AbstractWidget interface {
	// WidgetID returns the identifier of this widget within its parent container. Can be nil.
	WidgetID() ID
	// CreateElement creates the Element for this widget.
	// [Element] is used to represent and manage the widget in the UI tree.
	CreateElement(ctx *Context, parent Element) (Element, error)
}

// Widget represents a GUI widget laid out in a window.
type Widget interface {
	AbstractWidget
	ExclusiveType(marker.TypeWidget)
}

// AbstractContainer is an [AbstractWidget] that can contain child widgets.
type AbstractContainer interface {
	AbstractWidget
	// NumChildren returns the number of child widgets.
	NumChildren() int
	// Child returns the n-th child widget. Panics if n is out of range.
	Child(n int) AbstractWidget
	ExclusiveKind(marker.KindContainer)
}

// Container is a [Widget] that can contain child widgets.
type Container interface {
	AbstractContainer
	ExclusiveType(marker.TypeWidget)
}

// Menu represents a menu, such as window menu, context menu, or submenu.
// A Menu can contain MenuItems or wrap another Menu in which case it is
// a stateless or stateful menu.
type Menu interface {
	AbstractWidget
	ExclusiveType(marker.TypeMenu)
}

// MenuItem represents an item in a menu.
type MenuItem interface {
	AbstractWidget
	ExclusiveType(marker.TypeMenuItem)
}
