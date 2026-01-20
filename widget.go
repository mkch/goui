package goui

import "github.com/mkch/goui/marker"

// Component is the base interface for all widgets, menus and menu items.
type Component interface {
	// WidgetID returns the identifier of this widget within its parent container. Can be nil.
	WidgetID() ID
	// CreateElement creates the Element for this widget.
	// [Element] is used to represent and manage the widget in the UI tree.
	CreateElement(ctx *Context, parent Element) (Element, error)
}

// Widget represents a GUI widget laid out in a window.
type Widget interface {
	Component
	ExclusiveType(marker.TypeWidget)
}

// ContainerComponent is an [Component] that can contain child components.
type ContainerComponent interface {
	Component
	// NumChildren returns the number of child widgets.
	NumChildren() int
	// Child returns the n-th child widget. Panics if n is out of range.
	Child(n int) Component
	ExclusiveKind(marker.KindContainer)
}

// Container is a [Widget] that can contain child widgets.
type Container interface {
	ContainerComponent
	ExclusiveType(marker.TypeWidget)
}

// Menu represents a menu, such as window menu, context menu, or submenu.
// A Menu can contain MenuItems or wrap another Menu in which case it is
// a stateless or stateful menu.
type Menu interface {
	Component
	ExclusiveType(marker.TypeMenu)
}

// MenuItem represents an item in a menu.
type MenuItem interface {
	Component
	ExclusiveType(marker.TypeMenuItem)
}
