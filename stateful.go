package goui

import (
	"errors"
)

type StatefulWidget interface {
	Widget
	// CreateState creates the state associated with this widget.
	CreateState(*StateContext) State
	// Exclusive is a marker method to distinguish StatefulWidget, StatelessWidget and Container.
	Exclusive(StatefulWidget)
}

// StatefulWidgetHelper is a building block to implement [StatefulWidget].
// Embedding StatefulWidgetHelper in a struct and implementing WidgetID and
// CreateState methods allows the struct type to satisfy the [StatefulWidget] interface.
// See example in [StatefulWidget].
type StatefulWidgetHelper struct{}

func (StatefulWidgetHelper) Exclusive(StatefulWidget) { /*Nop*/ }

func (StatefulWidgetHelper) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx), nil
}

// StatefulWidgetFunc is a function type that implements [StatefulWidget].
// The WidgetID method returns nil.
// This function is convenient to create simple stateful widgets without defining a new struct type.
// See example of [StatefulWidget].
type StatefulWidgetFunc func(*StateContext) State

func (f StatefulWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatefulWidgetFunc) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (f StatefulWidgetFunc) CreateState(ctx *StateContext) State {
	return f(ctx)
}

func (f StatefulWidgetFunc) Exclusive(StatefulWidget) { /*Nop*/ }

type statefulElement struct {
	ElementBase
	state State
}

// Destroy implements [Element.Destroy].
func (e *statefulElement) Destroy() error {
	e.state.Destroy()
	return e.ElementBase.Destroy()
}

// createStatefulElement creates a new [Element] for a [StatefulWidget].
func createStatefulElement(*Context) Element {
	return &statefulElement{}
}

// State is the state associated with a [StatefulWidget].
type State interface {
	// Build builds the widget tree for this state.
	// It is called during the initial creation of the state
	// and whenever the state is updated via [UpdateStateFunc].
	Build() Widget
	// Destroy is called when the state is destroyed.
	// It can be used to clean up any resources associated with the state.
	Destroy()

	// Note: updater has little chance to yield an error,
	// because it typically just mutates a few simple fields in practice.
	// Having f return an error would add unnecessary noise for callers.

	// Update calls updater and then updates the state.
	// Use [StateUpdater] to implement this method.
	Update(updater func()) error
}

// simpleState is a simple implementation of State.
type simpleState struct {
	StateUpdater
	destroy func()        // Can be nil
	build   func() Widget // Can't be nil
}

func (s *simpleState) Build() Widget {
	return s.build()
}

func (s *simpleState) Destroy() {
	if s.destroy != nil {
		s.destroy()
	}
}

// NewState creates a new State with the given context, a build function, and a destroy function.
// The returned State uses the build function to build its widget tree, and the destroy function
// to clean up resources.
// The destroy func can be nil.
// This function is convenient to create simple states without defining a new struct type.
func NewState(ctx *StateContext, build func() Widget, destroy func()) State {
	return &simpleState{
		StateUpdater: NewStateUpdater(ctx),
		build:        build,
		destroy:      destroy,
	}
}

// NopDestroyer implements a no-op Destroy method.
// Embedding NopDestroyer in a struct provides a default implementation of Destroy method.
type NopDestroyer struct{}

// Destroy does nothing.
func (NopDestroyer) Destroy() { /*NOP*/ }

// StateUpdater implements State.Update method.
// Embedding StateUpdater in a struct and implementing other methods of [State]
// allows the struct type to satisfy the [State] interface.
type StateUpdater stateUpdater

type stateUpdater func(updater func()) error

// Update implements [State.Update].
func (h StateUpdater) Update(updater func()) error {
	return h(updater)
}

// StateContext is the context used by [NewStateUpdater].
type StateContext struct {
	*Context
	elem *statefulElement
}

// NewStateUpdater creates a new StateUpdater with the given context.
func NewStateUpdater(ctx *StateContext) StateUpdater {
	return func(updater func()) error {
		return updateWidgetState(updater, ctx.Context, ctx.elem)
	}
}

// updateWidgetState calls f and updates its widget tree.
// f can't be nil.
func updateWidgetState(f func(), ctx *Context, elem *statefulElement) error {
	f()
	// Rebuild the child widget and reconcile.
	newWidget := elem.state.Build()
	err := reconciledChildElement(ctx, elem, 0, newWidget)
	if err != nil {
		return err
	}
	layouter := layouterTree(elem.child(0))
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
	createState func(ctx *StateContext) State
}

func (w *statefulWidget) WidgetID() ID {
	return w.id
}

func (w *statefulWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (w *statefulWidget) CreateState(ctx *StateContext) State {
	return w.createState(ctx)
}

func (w *statefulWidget) Exclusive(StatefulWidget) { /*Nop*/ }

// NewStatefulWidget creates a new StatefulWidget with the given ID and createState function.
// The createState function is called in StatefulWidget.CreateState method.
func NewStatefulWidget(id ID, createState func(ctx *StateContext) State) StatefulWidget {
	return &statefulWidget{
		id:          id,
		createState: createState,
	}
}
