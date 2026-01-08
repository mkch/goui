package goui

import "github.com/mkch/goui/marker"

type WidgetBase interface {
	WidgetID() ID
	CreateElement(ctx *Context, parent Element) (Element, error)
}

type Widget interface {
	WidgetBase
	ExclusiveType(marker.TypeWidget)
}

type Menu interface {
	WidgetBase
	ExclusiveType(marker.TypeMenu)
}

type MenuItem interface {
	WidgetBase
	ExclusiveType(marker.TypeMenuItem)
}

type ContainerBase interface {
	WidgetBase
	NumChildren() int
	Child(n int) WidgetBase
	ExclusiveKind(marker.KindContainer)
}

type Container interface {
	ContainerBase
	ExclusiveType(marker.TypeWidget)
}
