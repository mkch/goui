package goui

type StatelessWidget interface {
	Widget
	Build(ctx *Context) Widget
}

type StatelessWidgetFunc func(ctx *Context) Widget

func (f StatelessWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatelessWidgetFunc) CreateElement(ctx *Context) (Element, error) {
	return CreateStatelessElement(ctx), nil
}

func (f StatelessWidgetFunc) Build(ctx *Context) Widget {
	return f(ctx)
}

// CreateStatelessElement creates a new [Element] for a [StatelessWidget].
func CreateStatelessElement(*Context) Element {
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
	return CreateStatelessElement(ctx), nil
}

func (w *statelessWidget) Build(ctx *Context) Widget {
	return w.build(ctx)
}

// NewStatelessWidget creates a new StatelessWidget with the given ID and build function.
// The build function is called in StatelessWidget.Build method.
func NewStatelessWidget(id ID, build func(ctx *Context) Widget) StatelessWidget {
	return &statelessWidget{
		id:    id,
		build: build,
	}
}
