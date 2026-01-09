package goui

import "github.com/mkch/goui/marker"

// AbstractStatelessWidget is a [AbstractWidget] that builds its UI using a Build method.
// Stateless widget is a wrapper around other widgets and do not hold any state.
type AbstractStatelessWidget interface {
	AbstractWidget
	// Build builds and returns the wrapped widget.
	Build(ctx *Context) AbstractWidget
	ExclusiveKind(marker.KindStateless)
}

// StatelessWidget is a [Widget] that builds its UI using a Build method.
type StatelessWidget interface {
	AbstractStatelessWidget
	ExclusiveType(marker.TypeWidget)
}

// StatelessHelper is a helper to implement concrete stateless widgets.
// The WidgetID method returns ID and the Build method calls Builder.
type StatelessHelper struct {
	ID
	// Builder is a function that builds and returns the wrapped widget. Can't be nil.
	Builder func(ctx *Context) AbstractWidget
}

// WidgetID implements [AbstractWidget.WidgetID].
func (w *StatelessHelper) WidgetID() ID {
	return w.ID
}

// CreateElement implements [AbstractWidget.CreateElement].
func (w *StatelessHelper) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatelessElement(ctx, parent)
}

// Build implements [AbstractStatelessWidget.Build].
// It calls the Builder function. Panics if Builder is nil.
func (w *StatelessHelper) Build(ctx *Context) AbstractWidget {
	return w.Builder(ctx)
}

func (*StatelessHelper) ExclusiveKind(marker.KindStateless) { /*Nop*/ }

// createStatelessElement creates and returns a new [Element] for a stateless widget.
func createStatelessElement(*Context, Element) (Element, error) {
	return &ElementHelper{}, nil
}

// statelessWidget is a [StatelessWidget] implementation.
type statelessWidget struct {
	StatelessHelper
}

func (*statelessWidget) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

// NewStatelessWidget creates a new [StatelessWidget] with the given ID and builder function.
func NewStatelessWidget(ID ID, builder func(ctx *Context) Widget) StatelessWidget {
	return &statelessWidget{
		StatelessHelper: StatelessHelper{
			ID:      ID,
			Builder: func(ctx *Context) AbstractWidget { return builder(ctx) },
		},
	}
}

// StatelessWidgetFunc is a function type that implements [StatelessWidget].
// Method WidgetID returns nil and Build calls f.
type StatelessWidgetFunc func(ctx *Context) Widget

// WidgetID implements [StatelessWidget.WidgetID].
func (f StatelessWidgetFunc) WidgetID() ID {
	return nil
}

// CreateElement implements [StatelessWidget.CreateElement].
func (f StatelessWidgetFunc) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatelessElement(ctx, parent)
}

// Build implements [StatelessWidget.Build].
// It calls f.
func (f StatelessWidgetFunc) Build(ctx *Context) AbstractWidget {
	return f(ctx)
}

func (StatelessWidgetFunc) ExclusiveType(marker.TypeWidget)    { /*Nop*/ }
func (StatelessWidgetFunc) ExclusiveKind(marker.KindStateless) { /*Nop*/ }
