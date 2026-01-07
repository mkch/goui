package goui

// StatelessWidget is a [Widget] that builds its UI using a builder function.
type StatelessWidget struct {
	ID
	Builder func(ctx *Context) Widget
}

func (w *StatelessWidget) WidgetID() ID {
	return w.ID
}

func (w *StatelessWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	return &ElementBase{}, nil
}

func (w *StatelessWidget) Build(ctx *Context) Widget {
	return w.Builder(ctx)
}

// Exclusive is a marker method to distinguish [StatelessWidget], [StatefulWidget] and [Container].
func (*StatelessWidget) Exclusive(StatelessWidget) { /*Nop*/ }
