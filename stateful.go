package goui

import (
	"errors"

	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/native"
)

// AbstractStatefulWidget is a [AbstractWidget] that has mutable state.
// The state is stored in a separate State object associated with the widget.
// The state can be updated via [State].Update method, which triggers a rebuild of the widget tree.
type AbstractStatefulWidget interface {
	AbstractWidget
	CreateState(ctx *StateContext) AbstractState
	ExclusiveKind(marker.KindStateful)
}

// StatefulWidgets is a [Widget] that has mutable state.
type StatefulWidget interface {
	AbstractStatefulWidget
	ExclusiveType(marker.TypeWidget)
}

// StatefulHelper is a helper to implement concrete state widgets.
type StatefulHelper struct {
	ID ID
	// StateCreator creates the state associated with this widget.
	StateCreator func(ctx *StateContext) AbstractState
}

func (w *StatefulHelper) WidgetID() ID {
	return w.ID
}

func (w *StatefulHelper) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx)
}

func (w *StatefulHelper) CreateState(ctx *StateContext) AbstractState {
	return w.StateCreator(ctx)
}

func (*StatefulHelper) ExclusiveKind(marker.KindStateful) { /*Nop*/ }

// statefulWidget is an implementation of [StatefulWidget].
type statefulWidget struct {
	StatefulHelper
}

func (*statefulWidget) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

// stateAdapter adapts a State to an [AbstractState].
type stateAdapter struct {
	State
}

func (a *stateAdapter) Build() AbstractWidget {
	return a.State.Build()
}

// NewStatefulWidget creates a new StatefulWidget with the given ID and state creator function.
func NewStatefulWidget(ID ID, stateCreator func(ctx *StateContext) State) StatefulWidget {
	return &statefulWidget{
		StatefulHelper: StatefulHelper{
			ID:           ID,
			StateCreator: func(ctx *StateContext) AbstractState { return &stateAdapter{State: stateCreator(ctx)} },
		},
	}
}

// StatefulWidgetFunc is a function type that implements [StatefulWidget].
// Method WidgetID returns nil and CreateState calls f.
type StatefulWidgetFunc func(ctx *StateContext) State

func (f StatefulWidgetFunc) WidgetID() ID {
	return nil
}

func (f StatefulWidgetFunc) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx)
}

func (f StatefulWidgetFunc) CreateState(ctx *StateContext) AbstractState {
	return &stateAdapter{State: f(ctx)}
}

func (StatefulWidgetFunc) ExclusiveKind(marker.KindStateful) { /*Nop*/ }

func (StatefulWidgetFunc) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

type statefulElement struct {
	ElementHelper
	state AbstractState
}

// Destroy implements [Element.Destroy].
func (e *statefulElement) Destroy() error {
	e.state.Destroy()
	return e.ElementHelper.Destroy()
}

// createStatefulElement creates a new [Element] for a [StatefulWidget].
func createStatefulElement(*Context) (Element, error) {
	return &statefulElement{}, nil
}

// AbstractState is the state associated with a [AbstractStatefulWidget].
type AbstractState interface {
	// Build builds the widget tree for this state.
	// It is called during the initial creation of the state
	// and whenever the state is updated via [UpdateStateFunc].
	Build() AbstractWidget
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

// State is the state associated with a [StatefulWidget].
type State interface {
	Build() Widget
	Destroy()
	Update(updater func()) error
}

// state is a simple implementation of [State].
type state struct {
	StateUpdater
	destroy func()        // Can be nil
	build   func() Widget // Can't be nil
}

func (s *state) Build() Widget {
	return s.build()
}

func (s *state) Destroy() {
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
	return &state{
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

	attachment := elementAttachedToWindow(ctx, statefulElement)
	if attachment == attachNone {
		return nil
	}

	if attachment == attachAsWindowMenu {
		return native.RefreshWindowMenu(ctx.NativeWindow())
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

type windowAttachment int

const (
	attachNone windowAttachment = iota
	attachToWidgetTree
	attachAsWindowMenu
)

// elementAttachedToWindow checks whether the given element is attached to the window.
// Returns [attachAsWindowMenu] only if element itself is the window menu or element's direct
// native parent is the window menu.
// Returns [attachToWidgetTree] if the element is part of the window's widget tree.
// Returns None otherwise.
func elementAttachedToWindow(ctx *Context, element Element) windowAttachment {
	// Check or window menu
	if _, isMenu := element.Widget().(interface{ ExclusiveType(marker.TypeMenu) }); isMenu {
		attachment, _ := LookupParent(element, func(e Element) (attachment windowAttachment, stop bool) {
			if e == ctx.window.Menu {
				return attachAsWindowMenu, true // Is window menu
			}
			widget := e.Widget()
			if _, ok := widget.(interface {
				ExclusiveType(marker.TypeMenu)
				ExclusiveKind(marker.KindStateless)
			}); ok {
				return attachNone, false // Continue searching
			}
			if _, ok := widget.(interface {
				ExclusiveType(marker.TypeMenu)
				ExclusiveKind(marker.KindStateful)
			}); ok {
				return attachNone, false // Continue searching
			}
			return attachNone, true // Not in window menu tree
		})
		return attachment
	}
	if _, isMenuitem := element.Widget().(interface{ ExclusiveType(marker.TypeMenuItem) }); isMenuitem {
		// MenuItem must have a parent, check it
		return elementAttachedToWindow(ctx, element.Parent())
	}
	// Check for window widget tree
	attachment, _ := LookupParent(element, func(e Element) (attachment windowAttachment, stop bool) {
		if e == ctx.window.Root {
			return attachToWidgetTree, true // In widget tree
		}
		return attachNone, false // Continue searching
	})
	return attachment
}

// errNotReplayable is returned when the parent layouter does not support replaying.
// errNotReplayable is a sentinel error.
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
