package goui

// StatelessWidgetBase is a [WidgetBase] that builds its UI using a builder function.
type StatelessWidgetBase interface {
	WidgetBase
	Build(ctx *Context) WidgetBase
	// Exclusive is a marker method to distinguish StatelessWidgetBase, StatefulWidgetBase and ContainerBase.
	Exclusive(StatelessWidgetBase)
}

// StatelessWidget is a [Widget] that builds its UI using a builder function.
type StatelessWidget interface {
	StatelessWidgetBase
	ExclusiveWidgetMenu(Widget)
}

// StatelessHelper is a helper to implement concrete stateless widgets.
type StatelessHelper struct {
	ID
	Builder func(ctx *Context) WidgetBase
}

func (w *StatelessHelper) WidgetID() ID {
	return w.ID
}

func (w *StatelessHelper) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatelessElement(ctx, parent)
}

func (w *StatelessHelper) Build(ctx *Context) WidgetBase {
	return w.Builder(ctx)
}

// Exclusive is a marker method to distinguish [StatelessWidget], [StatefulWidget] and [Container].
func (*StatelessHelper) Exclusive(StatelessWidgetBase) { /*Nop*/ }

func createStatelessElement(*Context, Element) (Element, error) {
	return &ElementBase{}, nil
}

type stateless struct {
	StatelessHelper
}

func (*stateless) ExclusiveWidgetMenu(Widget) { /*Nop*/ }

// NewStatelessWidget creates a new [StatelessWidget] with the given ID and builder function.
func NewStatelessWidget(ID ID, builder func(ctx *Context) Widget) StatelessWidget {
	return &stateless{
		StatelessHelper: StatelessHelper{
			ID:      ID,
			Builder: func(ctx *Context) WidgetBase { return builder(ctx) },
		},
	}
}

// StatelessWidgetFunc is a function type that implements [StatelessWidget].
// Method WidgetID returns nil and Build calls f.
type StatelessWidgetFunc func(ctx *Context) Widget

func (f StatelessWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatelessWidgetFunc) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatelessElement(ctx, parent)
}

func (f StatelessWidgetFunc) Build(ctx *Context) WidgetBase {
	return f(ctx)
}

func (StatelessWidgetFunc) Exclusive(StatelessWidgetBase) { /*Nop*/ }

func (StatelessWidgetFunc) ExclusiveWidgetMenu(Widget) { /*Nop*/ }
