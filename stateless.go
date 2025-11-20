package goui

type StatelessWidget interface {
	Widget
	Build(ctx *Context) Widget
	// Exclusive is a marker method to distinguish StatefulWidget, StatelessWidget and Container.
	Exclusive(statelessWidget)
}

// StatelessWidgetImpl is a building block to implement [StatelessWidget].
// Embedding this struct in a struct and implementing the remaining methods of
// [StatelessWidget] allows the struct type to satisfy the [StatelessWidget] interface.
type StatelessWidgetImpl struct{}

func (StatelessWidgetImpl) Exclusive(statelessWidget) { /*Nop*/ }

func (StatelessWidgetImpl) CreateElement(ctx *Context) (Element, error) {
	return createStatelessElement(ctx), nil
}

type StatelessWidgetFunc func(ctx *Context) Widget

func (f StatelessWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatelessWidgetFunc) CreateElement(ctx *Context) (Element, error) {
	return createStatelessElement(ctx), nil
}

func (f StatelessWidgetFunc) Build(ctx *Context) Widget {
	return f(ctx)
}

func (f StatelessWidgetFunc) Exclusive(statelessWidget) { /*Nop*/ }

// createStatelessElement creates a new [Element] for a [StatelessWidget].
func createStatelessElement(*Context) Element {
	return &ElementBase{}
}

// statelessWidget is an implementation of StatelessWidget.
type statelessWidget struct {
	id    ID
	build func(ctx *Context) Widget
}

func (w *statelessWidget) WidgetID() ID {
	return w.id
}

func (w *statelessWidget) CreateElement(ctx *Context) (Element, error) {
	return createStatelessElement(ctx), nil
}

func (w *statelessWidget) Build(ctx *Context) Widget {
	return w.build(ctx)
}

func (w *statelessWidget) Exclusive(statelessWidget) { /*Nop*/ }

// NewStatelessWidget creates a new StatelessWidget with the given ID and build function.
// The build function is called in StatelessWidget.Build method.
func NewStatelessWidget(id ID, build func(ctx *Context) Widget) StatelessWidget {
	return &statelessWidget{
		id:    id,
		build: build,
	}
}
