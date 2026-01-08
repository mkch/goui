package goui

import (
	"errors"
	"fmt"

	"github.com/mkch/goui/native"
)

// StatefulWidgetBase is a [WidgetBase] that has mutable state.
// The state is stored in a separate State object associated with the widget.
// The state can be updated via [State].Update method, which triggers a rebuild of the widget tree.
type StatefulWidgetBase interface {
	WidgetBase
	CreateState(ctx *StateContext) StateBase
	// Exclusive is a marker method to distinguish StatefulWidgetBase, StatelessWidgetBase and ContainerBase.
	Exclusive(StatefulWidgetBase)
}

// StatefulWidgets is a [Widget] that has mutable state.
type StatefulWidget interface {
	StatefulWidgetBase
	ExclusiveWidgetMenu(Widget)
}

// StatefulHelper is a helper to implement concrete state widgets.
type StatefulHelper struct {
	ID ID
	// StateCreator creates the state associated with this widget.
	StateCreator func(ctx *StateContext) StateBase
}

func (w *StatefulHelper) WidgetID() ID {
	return w.ID
}

func (w *StatefulHelper) CreateElement(ctx *Context, parent Element) (Element, error) {
	return createStatefulElement(ctx)
}

func (w *StatefulHelper) CreateState(ctx *StateContext) StateBase {
	return w.StateCreator(ctx)
}

func (*StatefulHelper) Exclusive(StatefulWidgetBase) { /*Nop*/ }

type stateful struct {
	StatefulHelper
}

func (*stateful) ExclusiveWidgetMenu(Widget) { /*Nop*/ }

// fromStateAdapter adapts a State to an StateBase.
type fromStateAdapter struct {
	State
}

func (a *fromStateAdapter) Build() WidgetBase {
	return a.State.Build()
}

// NewStatefulWidget creates a new StatefulWidget with the given ID and state creator function.
func NewStatefulWidget(ID ID, stateCreator func(ctx *StateContext) State) StatefulWidget {
	return &stateful{
		StatefulHelper: StatefulHelper{
			ID:           ID,
			StateCreator: func(ctx *StateContext) StateBase { return &fromStateAdapter{State: stateCreator(ctx)} },
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

func (f StatefulWidgetFunc) CreateState(ctx *StateContext) StateBase {
	return &fromStateAdapter{State: f(ctx)}
}

func (f StatefulWidgetFunc) Exclusive(StatefulWidgetBase) { /*Nop*/ }

func (f StatefulWidgetFunc) ExclusiveWidgetMenu(Widget) { /*Nop*/ }

type statefulElement struct {
	ElementBase
	state StateBase
}

// Destroy implements [Element.Destroy].
func (e *statefulElement) Destroy() error {
	e.state.Destroy()
	return e.ElementBase.Destroy()
}

// createStatefulElement creates a new [Element] for a [StatefulWidget].
func createStatefulElement(*Context) (Element, error) {
	return &statefulElement{}, nil
}

// StateBase is the state associated with a [StatefulWidgetBase].
type StateBase interface {
	// Build builds the widget tree for this state.
	// It is called during the initial creation of the state
	// and whenever the state is updated via [UpdateStateFunc].
	Build() WidgetBase
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
	if attachment == None {
		return nil
	}

	if attachment == IsWindowMenu {
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
	None windowAttachment = iota
	InWidgetTree
	IsWindowMenu
)

// elementAttachedToWindow checks whether the given element is attached to the window.
// Returns IsWindowMenu only if element itself is the window menu or element's direct
// native parent is the window menu.
// Returns InWidgetTree if the element is in the window's main widget tree.
// Returns None otherwise.
func elementAttachedToWindow(ctx *Context, element Element) windowAttachment {
	const (
		stateInitial = iota
		stateInWidgetTree
		stateInMenuTree
	)

	state := stateInitial

	for elem := element; elem != nil; elem = elem.Parent() {
		switch state {
		case stateInitial:
			// Check for window menu or root
			if elem == ctx.window.Menu {
				return IsWindowMenu
			}
			if elem == ctx.window.Root {
				return InWidgetTree
			}

			if _, ok := elem.(NativeMenuElement); ok {
				// Non-window menu, not attached to window
				return None
			}

			// Check for menu item, enter menu tree state
			if _, ok := elem.(NativeMenuItemElement); ok {
				// May be an item of the window menu
				state = stateInMenuTree
				continue
			}

			// Skip transparent widgets (StatefulWidget and StatelessWidget)
			widget := elem.Widget()
			if _, ok := widget.(StatefulWidget); ok {
				continue
			}
			if _, ok := widget.(StatelessWidget); ok {
				continue
			}

			// Encountered non-transparent widget (Container or other), enter widget tree state
			state = stateInWidgetTree

		case stateInWidgetTree:
			// Check for root widget
			if elem == ctx.window.Root {
				return InWidgetTree
			}

			// Skip transparent widgets and containers
			widget := elem.Widget()
			if _, ok := widget.(StatefulWidget); ok {
				continue
			}
			if _, ok := widget.(StatelessWidget); ok {
				continue
			}
			if _, ok := widget.(Container); ok {
				continue
			}

			if ctx.app.debug != nil {
				if _, ok := widget.(NativeMenuElement); ok {
					panic("invalid element tree: widget tree contains menu element")
				}
				if _, ok := widget.(NativeMenuItemElement); ok {
					panic("invalid element tree: widget tree contains menu item element")
				}
			}

			// Encountered other element, not in window widget tree
			return None

		case stateInMenuTree:
			// Check for window menu
			if elem == ctx.window.Menu {
				return IsWindowMenu
			}

			if _, ok := elem.(NativeMenuElement); ok {
				// Non-window menu, not attached to window
				return None
			}

			// Skip transparent widgets
			widget := elem.Widget()
			if _, ok := widget.(StatefulWidget); ok {
				continue
			}
			if _, ok := widget.(StatelessWidget); ok {
				continue
			}

			// If it's neither transparent widget nor menu, the element tree
			// is illegal.
			if ctx.app.debug != nil {
				panic(fmt.Sprintf("invalid element tree: menu tree contains non-menu element %T", elem))
			}
			return None
		}
	}

	// Reached end of parent chain without finding root or window menu
	return None
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
