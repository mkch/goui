package goui

import (
	"errors"
)

type StatefulWidget interface {
	Widget
	CreateState(*Context) *WidgetState
	// Exclusive is a marker method to distinguish StatefulWidget, StatelessWidget and Container.
	Exclusive(StatefulWidget)
}

// StatefulWidgetImpl is a building block to implement [StatefulWidget].
// Embedding StatefulWidgetImpl in a struct and implementing the remaining methods of
// [StatefulWidget] allows the struct type to satisfy the [StatefulWidget] interface.
type StatefulWidgetImpl struct{}

func (StatefulWidgetImpl) Exclusive(StatefulWidget) { /*Nop*/ }

func (StatefulWidgetImpl) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx), nil
}

type StatefulWidgetFunc func(*Context) *WidgetState

func (f StatefulWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatefulWidgetFunc) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (f StatefulWidgetFunc) CreateState(ctx *Context) *WidgetState {
	return f(ctx)
}

func (f StatefulWidgetFunc) Exclusive(StatefulWidget) { /*Nop*/ }

type statefulElement struct {
	ElementBase
	state *WidgetState
}

func (e *statefulElement) destroy() {
	e.state.destroyData()
	e.ElementBase.destroy()
}

// createStatefulElement creates a new [Element] for a [StatefulWidget].
func createStatefulElement(*Context) Element {
	return &statefulElement{}
}

type WidgetState struct {
	// Build builds the widget tree for this state.
	// It is called during the initial creation of the state
	// and whenever the state is updated via WidgetState.Update.
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

	elem, layouter, err := updateElementTree(ws.ctx, ws.element, ws.element.Widget())
	if err != nil {
		return err
	}
	if ws.element == ws.ctx.window.Root {
		ws.ctx.window.Root = elem
		ws.ctx.window.Layouter = layouter
	}

	if layouter == nil {
		return nil
	}

	if err = replayParentLayouter(ws.ctx, layouter); err == nil {
		return nil
	}

	if err == errNotReplayable {
		return layoutWindow(ws.ctx.window)
	}
	return err
}

var errNotReplayable = errors.New("the parent layouter does not support replaying")

// replayParentLayouter replays the laying out of the nearest recursive parent
// which supports replaying.
// If no such parent exists, it returns errNotReplayable.
func replayParentLayouter(ctx *Context, root Layouter) error {
	// Find the nearest child-independent recursive parent(replayer).
	var replayer func(*Context) error
	for parent := root.parent(); parent != nil; parent = parent.parent() {
		if replayer = parent.Replayer(); replayer != nil {
			break
		}
	}
	if replayer == nil {
		return errNotReplayable
	}
	return replayer(ctx)
}

// statelessWidget is an implementation of StatelessWidget.
type statefulWidget struct {
	id          ID
	createState func(ctx *Context) *WidgetState
}

func (w *statefulWidget) WidgetID() ID {
	return w.id
}

func (w *statefulWidget) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (w *statefulWidget) CreateState(ctx *Context) *WidgetState {
	return w.createState(ctx)
}

func (w *statefulWidget) Exclusive(StatefulWidget) { /*Nop*/ }

// NewStatefulWidget creates a new StatefulWidget with the given ID and createState function.
// The createState function is called in StatefulWidget.CreateState method.
func NewStatefulWidget(id ID, createState func(ctx *Context) *WidgetState) StatefulWidget {
	return &statefulWidget{
		id:          id,
		createState: createState,
	}
}
