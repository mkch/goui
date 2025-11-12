package goui

type StatefulWidget interface {
	Widget
	CreateState(*Context) *WidgetState
}

type StatefulWidgetFuc func(*Context) *WidgetState

func (f StatefulWidgetFuc) WidgetID() ID {
	return nil
}

func (f StatefulWidgetFuc) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx)
}

func (f StatefulWidgetFuc) CreateState(ctx *Context) *WidgetState {
	return f(ctx)
}

type Stateful struct{}

func (s Stateful) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx)
}

type statefulElement struct {
	element
	state *WidgetState
}

func (e *statefulElement) destroy() {
	e.state.destroyData()
	e.element.destroy()
}

func createStatefulElement(ctx *Context) (Element, error) {
	return &statefulElement{}, nil
}

type WidgetState struct {
	// Build builds the widget tree for this state.
	// It is called during the initial creation of the state
	// and whenever the state is updated via [WidgetState.Update].
	Build func() Widget
	// DestroyData is called when the state is destroyed, if not nil.
	// It can be used to clean up any resources associated with the state.
	DestroyData func() // Can be nil

	ctx     *Context
	element Element
}

func (ws *WidgetState) destroyData() {
	if ws.DestroyData != nil {
		ws.DestroyData()
	}
}

func (ws *WidgetState) Update(updater func()) error {
	updater()

	err := updateElementTree(ws.ctx, ws.element, ws.element.widget())
	if err != nil {
		return err
	}
	layouter, err := buildLayouterTree(ws.ctx, ws.ctx.window.Root)
	if err != nil {
		return err
	}

	ws.ctx.window.Layouter = layouter

	return layoutWindow(ws.ctx.window)
}

// statelessWidget is an implementation of StatelessWidget.
type statefulWidget struct {
	Stateful
	id          ID
	createState func(ctx *Context) *WidgetState
}

func (w *statefulWidget) WidgetID() ID {
	return w.id
}

func (w *statefulWidget) CreateState(ctx *Context) *WidgetState {
	return w.createState(ctx)
}

// NewStatefulWidget creates a new StatefulWidget with the given ID and createState function.
// The createState function is called in [StatefulWidget.CreateState] method.
func NewStatefulWidget(id ID, createState func(ctx *Context) *WidgetState) StatefulWidget {
	return &statefulWidget{
		id:          id,
		createState: createState,
	}
}
