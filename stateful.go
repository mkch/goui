package goui

import (
	"errors"

	"github.com/mkch/goui/native"
)

// StatefulWidget is a widget that has mutable state.
// The state is stored in a separate State object associated with the widget.
// The state can be updated via [State].Update method, which triggers a rebuild of the widget tree.
type StatefulWidget struct {
	ID ID
	// StateCreator creates the state associated with this widget.
	StateCreator func(ctx *StateContext) State
}

func (w *StatefulWidget) WidgetID() ID {
	return w.ID
}

func (w *StatefulWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx), nil
}

func (w *StatefulWidget) CreateState(ctx *StateContext) State {
	return w.StateCreator(ctx)
}

// Exclusive is a marker method to distinguish StatefulWidget, StatelessWidget and Container.
func (w *StatefulWidget) Exclusive(StatefulWidget) { /*Nop*/ }

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
	// Use StateUpdater to implement this method.
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

// StateUpdater implements [State].Update method.
// Embedding StateUpdater in a struct and implementing other methods of [State]
// allows the struct type to satisfy the [State] interface.
type StateUpdater stateUpdater

type stateUpdater func(updater func()) error

// Update implements [State].Update.
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
func updateWidgetState(f func(), ctx *Context, statefulElement *statefulElement) error {
	f()

	// Rebuild the child widget and reconcile.
	newWidget := statefulElement.state.Build()
	err := reconciledChildElement(ctx, statefulElement, 0, newWidget)
	if err != nil {
		return err
	}

	// Whether statefulElement is part of the window element tree.
	var rootedInWindow bool
	// If the updated stateful element is part of a menu, refresh the menu.
	for elem := Element(statefulElement); elem != nil; elem = elem.Parent() {
		if elem == ctx.window.Menu {
			if err := native.RefreshWindowMenu(ctx.NativeWindow()); err != nil {
				return err
			}
			break
		}
		if elem == ctx.window.Root {
			rootedInWindow = true
			break
		}
	}

	if !rootedInWindow {
		return nil
	}

	// Layout
	layouter := layouterTree(statefulElement.Child(0))
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
