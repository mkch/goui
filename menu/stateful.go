package menu

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
)

// StatefulMenu is a stateful wrapper for a [Menu].
type StatefulMenu interface {
	goui.StatefulComponent
	ExclusiveType(marker.TypeMenu)
}

// statefulMenu is a [StatefulMenu] implementation.
type statefulMenu struct {
	goui.StatefulHelper
}

func (*statefulMenu) ExclusiveType(marker.TypeMenu) { /*Nop*/ }

// MenuState is the state associated with a [StatefulMenu].
type MenuState interface {
	Build() goui.Menu
	Destroy()
	Update(updater func()) error
}

// menuState is an implementation of MenuState.
type menuState struct {
	stateHelper
	builder func() goui.Menu // Can't be nil
}

func (s *menuState) Build() goui.Menu {
	return s.builder()
}

// NewMenuState creates a new MenuState using the given build and destroy functions.
func NewMenuState(ctx *goui.StateContext, builder func() goui.Menu, destroyer func()) MenuState {
	return &menuState{
		stateHelper: stateHelper{
			StateUpdater: goui.NewStateUpdater(ctx),
			destroy:      destroyer,
		},
		builder: builder,
	}
}

// menuStateAdapter adapts a MenuState to a goui.State.
type menuStateAdapter struct {
	MenuState
}

func (a menuStateAdapter) Build() goui.Component {
	return a.MenuState.Build()
}

// NewStatefulMenu creates a new StatefulMenu with the given ID and state creator function.
func NewStatefulMenu(ID goui.ID, stateCreator func(ctx *goui.StateContext) MenuState) StatefulMenu {
	return &statefulMenu{
		StatefulHelper: goui.StatefulHelper{
			ID:           ID,
			StateCreator: func(ctx *goui.StateContext) goui.ComponentState { return &menuStateAdapter{stateCreator(ctx)} },
		},
	}
}

// StatefulMenuFunc is a function type that implements [StatefulMenu].
type StatefulMenuFunc func(ctx *goui.StateContext) MenuState

// WidgetID implements [StatefulMenu.WidgetID] and always returns nil.
func (f StatefulMenuFunc) WidgetID() goui.ID {
	return nil
}

// CreateElement implements [StatefulMenu.CreateElement].
func (f StatefulMenuFunc) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.StatefulElement{}, nil
}

// CreateState implements [StatefulMenu.CreateState] and calls f(ctx) to build the state.
func (f StatefulMenuFunc) CreateState(ctx *goui.StateContext) goui.ComponentState {
	return &menuStateAdapter{f(ctx)}
}

func (StatefulMenuFunc) ExclusiveType(marker.TypeMenu)     { /*Nop*/ }
func (StatefulMenuFunc) ExclusiveKind(marker.KindStateful) { /*Nop*/ }

type StatefulItem interface {
	goui.StatefulComponent
	ExclusiveType(marker.TypeMenuItem)
}

type statefulItem struct {
	goui.StatefulHelper
}

func (*statefulItem) ExclusiveType(marker.TypeMenuItem) { /*Nop*/ }

// ItemState is the state associated with a [StatefulItem].
// See [goui.WidgetState] for more details.
type ItemState interface {
	Build() goui.MenuItem
	Destroy()
	Update(updater func()) error
}

// stateHelper is the building block for [menuState] and [itemState].
type stateHelper struct {
	goui.StateUpdater
	destroy func() // Can be nil
}

func (s *stateHelper) Destroy() {
	if s.destroy != nil {
		s.destroy()
	}
}

// itemState is an implementation of ItemState.
type itemState struct {
	stateHelper
	build func() goui.MenuItem // Can't be nil
}

func (s *itemState) Build() goui.MenuItem {
	return s.build()
}

func NewItemState(ctx *goui.StateContext, build func() goui.MenuItem, destroy func()) ItemState {
	return &itemState{
		stateHelper: stateHelper{
			StateUpdater: goui.NewStateUpdater(ctx),
			destroy:      destroy,
		},
		build: build,
	}
}

// itemStateAdapter adapts an ItemState to a goui.State.
type itemStateAdapter struct {
	ItemState
}

func (a itemStateAdapter) Build() goui.Component {
	return a.ItemState.Build()
}

// NewStatefulMenuItem creates a new StatefulMenuItem with the given ID and state creator function.
func NewStatefulItem(ID goui.ID, stateCreator func(ctx *goui.StateContext) ItemState) StatefulItem {
	return &statefulItem{
		StatefulHelper: goui.StatefulHelper{
			ID:           ID,
			StateCreator: func(ctx *goui.StateContext) goui.ComponentState { return &itemStateAdapter{stateCreator(ctx)} },
		},
	}
}

// StatefulItemFunc is a function type that implements [StatefulItem].
type StatefulItemFunc func(ctx *goui.StateContext) ItemState

// WidgetID implements [StatefulItem.WidgetID] and always returns nil.
func (f StatefulItemFunc) WidgetID() goui.ID {
	return nil
}

// CreateElement implements [StatefulItem.CreateElement].
func (f StatefulItemFunc) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.StatefulElement{}, nil
}

// CreateState implements [StatefulItem.CreateState] and calls f(ctx) to build the state.
func (f StatefulItemFunc) CreateState(ctx *goui.StateContext) goui.ComponentState {
	return &itemStateAdapter{f(ctx)}
}

func (StatefulItemFunc) ExclusiveType(marker.TypeMenuItem) { /*Nop*/ }
func (StatefulItemFunc) ExclusiveKind(marker.KindStateful) { /*Nop*/ }
