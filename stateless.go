package goui

// StatelessWidget is a [Widget] that builds its UI using a builder function.
type StatelessWidget interface {
	Widget
	Build(ctx *Context) Widget
	// Exclusive is a marker method to distinguish StatelessWidget, StatefulWidget and Container.
	Exclusive(StatelessWidget)
}

// Stateless implements a [StatelessWidget].
type Stateless struct {
	ID
	Builder func(ctx *Context) Widget
}

func (w *Stateless) WidgetID() ID {
	return w.ID
}

func (w *Stateless) CreateElement(ctx *Context, parent Element) (Element, error) {
	return &ElementBase{}, nil
}

func (w *Stateless) Build(ctx *Context) Widget {
	return w.Builder(ctx)
}

// Exclusive is a marker method to distinguish [StatelessWidget], [StatefulWidget] and [Container].
func (*Stateless) Exclusive(StatelessWidget) { /*Nop*/ }
