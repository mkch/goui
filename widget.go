package goui

type WidgetBase interface {
	WidgetID() ID
	CreateElement(ctx *Context, parent Element) (Element, error)
}

type Widget interface {
	WidgetBase
	// Exclusive is a marker method to distinguish [Widget], [Menu] and [MenuItem].
	ExclusiveWidgetMenu(Widget)
}

type Menu interface {
	WidgetBase
	// Exclusive is a marker method to distinguish [Widget], [Menu] and [MenuItem].
	ExclusiveWidgetMenu(Menu)
}

type MenuItem interface {
	WidgetBase
	// Exclusive is a marker method to distinguish [Widget], [Menu] and [MenuItem].
	ExclusiveWidgetMenu(MenuItem)
}

type ContainerBase interface {
	WidgetBase
	NumChildren() int
	Child(n int) WidgetBase
	// Exclusive is a marker method to distinguish StatefulWidgetBase, StatelessWidgetBase and ContainerBase.
	Exclusive(ContainerBase)
}

type Container interface {
	ContainerBase
	ExclusiveWidgetMenu(Widget)
}
