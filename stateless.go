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
	return createStatelessElement(ctx)
}

func (f StatelessWidgetFunc) Build(ctx *Context) Widget {
	return f(ctx)
}

type Stateless struct{}

func (s Stateless) CreateElement(ctx *Context) (Element, error) {
	return createStatelessElement(ctx)
}

func createStatelessElement(*Context) (Element, error) {
	return &element{}, nil
}

// statelessWidget is an implementation of StatelessWidget.
type statelessWidget struct {
	Stateless
	id    ID
	build func(ctx *Context) Widget
}

func (w *statelessWidget) WidgetID() ID {
	return w.id
}

func (w *statelessWidget) Build(ctx *Context) Widget {
	return w.build(ctx)
}

// NewStatelessWidget creates a new StatelessWidget with the given ID and build function.
// The build function is called in [StatelessWidget.Build] method.
func NewStatelessWidget(id ID, build func(ctx *Context) Widget) StatelessWidget {
	return &statelessWidget{
		id:    id,
		build: build,
	}
}
