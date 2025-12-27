package goui

import (
	"errors"
)

type StatefulWidget interface {
	Widget
	CreateState(*Context, UpdateStateFunc) *WidgetState
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

type StatefulWidgetFunc func(*Context, UpdateStateFunc) *WidgetState

func (f StatefulWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatefulWidgetFunc) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (f StatefulWidgetFunc) CreateState(ctx *Context, updateState UpdateStateFunc) *WidgetState {
	return f(ctx, updateState)
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

// Note: Parameter f in [UpdateStateFunc] has little chance to yield an error,
// because it typically just mutates a few simple fields in practice.
// Having f return an error would add unnecessary noise for callers.

// UpdateStateFunc is a function type for updating the state of a widget.
// Parameter f is a function that performs the state update.
// UpdateStateFunc calls f and updates the widget accordingly.
type UpdateStateFunc func(f func()) error

type WidgetState struct {
	// Build builds the widget tree for this state.
	// It is called during the initial creation of the state
	// and whenever the state is updated via [UpdateStateFunc].
	Build func() Widget
	// DestroyData is called when the state is destroyed, if not nil.
	// It can be used to clean up any resources associated with the state.
	DestroyData func() // Can be nil
}

func (ws *WidgetState) destroyData() {
	if ws.DestroyData != nil {
		ws.DestroyData()
	}
}

// updateWidgetState calls f and updates its widget tree.
// f can't be nil.
func updateWidgetState(f func(), ctx *Context, elem *statefulElement) error {
	f()
	// Rebuild the child widget and reconcile.
	newWidget := elem.state.Build()
	reconciled, layouter, err := reconcileElementTree(ctx, elem.children[0], newWidget)
	if err != nil {
		return err
	}
	if reconciled != elem.children[0] {
		element_SetChild(elem, 0, reconciled)
	}

	if layouter == nil {
		return nil
	}

	if err = replayParentLayouter(ctx, layouter); err == nil {
		return nil
	}

	if err == errNotReplayable {
		return layoutWindow(ctx)
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
	for parent := root.Parent(); parent != nil; parent = parent.Parent() {
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
	createState func(ctx *Context, updateState UpdateStateFunc) *WidgetState
}

func (w *statefulWidget) WidgetID() ID {
	return w.id
}

func (w *statefulWidget) CreateElement(ctx *Context) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (w *statefulWidget) CreateState(ctx *Context, updateState UpdateStateFunc) *WidgetState {
	return w.createState(ctx, updateState)
}

func (w *statefulWidget) Exclusive(StatefulWidget) { /*Nop*/ }

// NewStatefulWidget creates a new StatefulWidget with the given ID and createState function.
// The createState function is called in StatefulWidget.CreateState method.
func NewStatefulWidget(id ID, createState func(ctx *Context, updateState UpdateStateFunc) *WidgetState) StatefulWidget {
	return &statefulWidget{
		id:          id,
		createState: createState,
	}
}
