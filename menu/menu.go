// Package menu implements menu widgets.
package menu

import (
	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// Menu is a menu widget that can contain items of type [Item] and [Separator].
type Menu struct {
	ID goui.ID
	// Items are the menu's child widgets.
	// Only widgets representing menu items should be added;
	// others will not display correctly or cause error returns.
	Items []goui.MenuItem
}

// WidgetID implements [goui.Widget.WidgetID].
func (m *Menu) WidgetID() goui.ID {
	return m.ID
}

// CreateElement implements [goui.Widget.CreateElement].
func (m *Menu) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	return createMenuElement(parent, true)
}

func (*Menu) ExclusiveWidgetMenu(goui.Menu) { /*Nop*/ }

// lookupNativeItemParent searches the element and its ancestors for the nearest [*nativeItemElement]
// and returns its handle, or nil if not found.
// The search stops and returns ErrWrongParent error when encountering an element whose widget is neither
// [goui.StatelessWidget], [goui.StatefulWidget] nor [*nativeItemElement].
func lookupNativeItemParent(element goui.Element) (native.Handle, error) {
	item, err := lookupNativeParent[goui.NativeMenuItemElement, StatelessMenu, StatefulMenu](element)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return item.NativeMenuItem(), nil
}

// lookupNativeMenuParent searches the element and its ancestors for the nearest [*nativeMenuElement]
// and returns its handle, or nil if not found.
// The search stops and returns ErrWrongParent error when encountering an element whose widget is neither
// [goui.StatelessWidget], [goui.StatefulWidget] nor [*nativeMenuElement].
func lookupNativeMenuParent(element goui.Element) (native.Handle, error) {
	menu, err := lookupNativeParent[goui.NativeMenuElement, StatelessItem, StatefulItem](element)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, nil
	}
	return menu.NativeMenu(), nil
}

// lookupNativeParent is a helper function for [lookupNativeMenuParent] and [lookupNativeItemParent].
// It searches the element and its ancestors for the nearest element of type T
// and returns it via ret, or nil if not found.
// The search stops and returns ErrWrongParent error when encountering an element whose widget is neither
// StatelessType, StatefulType nor T.
// StatelessType and StatefulType should be the stateless and stateful version of counter T.
func lookupNativeParent[
	ElementType goui.Element,
	StatelessType goui.StatelessWidgetBase,
	StatefulType goui.StatefulWidgetBase,
](element goui.Element) (ret ElementType, err error) {
	type R struct {
		val ElementType
		err error
	}
	r, found := goui.LookupParent(element, func(e goui.Element) (R, bool) {
		if elem, ok := e.(ElementType); ok {
			return R{elem, nil}, true
		}
		widget := e.Widget()
		if _, isStateless := widget.(StatelessType); isStateless {
			return R{}, false // Continue searching
		}
		if _, isStateful := widget.(StatefulType); isStateful {
			return R{}, false // Continue searching
		}
		var zero ElementType
		return R{zero, errortrace.WithStack(goui.ErrWrongParent)}, true // Wrong parent
	})
	if found {
		ret = r.val
		err = r.err
	}
	return
}

// createMenuElement is a helper function to create a popup or window menu element.
func createMenuElement(parent goui.Element, popup bool) (element goui.Element, err error) {
	handle := native.CreateMenu(popup)
	if parent != nil {
		// If parent is given, it must represent a opener
		// menu item to which this menu is a submenu.
		var opener native.Handle
		if opener, err = lookupNativeItemParent(parent); err != nil {
			err = errortrace.ErrorfStack("create menu failed: %w", err)
			return
		}
		if opener != nil {
			if err = native.SetMenuItemSubmenu(opener, handle); err != nil {
				return
			}
		}
	}
	return &nativeMenuElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}, nil
}

// NumChildren implements [goui.Container.NumChildren].
func (m *Menu) NumChildren() int {
	return len(m.Items)
}

// Child implements [goui.Container.Child].
func (m *Menu) Child(index int) goui.WidgetBase {
	return m.Items[index]
}

// Exclusive implements [goui.Container.Exclusive].
func (m *Menu) Exclusive(goui.ContainerBase) { /*Nop*/ }

// WindowMenu is a [Menu] that is suitable for use as a window menu.
// Use this type for window menus and [Menu] for submenus/popups;
// mixing them incorrectly may cause platform-specific issues.
type WindowMenu Menu

// WidgetID implements [goui.Widget.WidgetID].
func (m *WindowMenu) WidgetID() goui.ID {
	return m.ID
}

// CreateElement implements [goui.Widget.CreateElement].
func (m *WindowMenu) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return createMenuElement(parent, false)
}

// NumChildren implements [goui.Container.NumChildren].
func (m *WindowMenu) NumChildren() int {
	return len(m.Items)
}

// Child implements [goui.Container.Child].
func (m *WindowMenu) Child(index int) goui.WidgetBase {
	return m.Items[index]
}

// Exclusive implements [goui.Container.Exclusive].
func (m *WindowMenu) Exclusive(goui.ContainerBase) { /*Nop*/ }

// Exclusive implements [goui.Widget.ExclusiveWidgetMenu].
func (*WindowMenu) ExclusiveWidgetMenu(goui.Menu) { /*Nop*/ }

// nativeMenuElement is an implementation of [goui.NativeMenuElement]
// that represents a native menu.
type nativeMenuElement struct {
	goui.ElementBase
	Handle native.Handle
}

// NativeMenu implements [goui.NativeMenuElement.NativeMenu].
func (e *nativeMenuElement) NativeMenu() native.Handle {
	return e.Handle
}

// Destroy implements [goui.Element.Destroy].
func (e *nativeMenuElement) Destroy() (err error) {
	return native.DestroyMenu(e.Handle)
}

// Item is a menu item widget that can be added to a [Menu].
type Item struct {
	ID goui.ID
	// Submenu is an optional submenu of this menu item.
	// Only widgets representing menus should be assigned;
	// others will not display correctly or cause error returns.
	Submenu  goui.Menu
	Title    string
	Disabled bool
	OnSelect func(*goui.Context)
}

// WidgetID implements [goui.Widget.WidgetID].
func (item *Item) WidgetID() goui.ID {
	return item.ID
}

// NumChildren implements [goui.Container.NumChildren].
func (item *Item) NumChildren() int {
	return gg.If(item.Submenu != nil, 1, 0)
}

// Child implements [goui.Container.Child].
func (item *Item) Child(index int) goui.WidgetBase {
	if item.Submenu == nil || index != 0 {
		panic("index out of range")
	}
	return item.Submenu
}

// Exclusive implements [goui.Container.Exclusive].
func (item *Item) Exclusive(goui.ContainerBase) { /*Nop*/ }

// Exclusive implements [goui.Widget.ExclusiveWidgetMenu].
func (*Item) ExclusiveWidgetMenu(goui.MenuItem) { /*Nop*/ }

// CreateElement implements [goui.Widget.CreateElement].
func (item *Item) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	handle, err := createItem(parent, item.Title, false)
	if err != nil {
		return
	}
	element = &nativeItemElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}
	return
}

// createItem is a helper function to create a menu item under the given parent element.
// If separator is true, a separator item is created.
func createItem(parent goui.Element, title string, separator bool) (handle native.Handle, err error) {
	parentMenu, err := lookupNativeMenuParent(parent)
	if err == nil && parentMenu == nil {
		// Menu item must have a parent menu
		err = goui.ErrWrongParent
	}
	if err != nil {
		err = errortrace.ErrorfStack("create menu %s failed: %w", gg.If(separator, "separator", "item"), err)
		return
	}
	handle, err = native.CreateMenuItem(parentMenu, title, separator)
	return
}

// nativeItemElement is an implementation of [goui.NativeMenuItemElement]
// that represents a native menu item.
type nativeItemElement struct {
	goui.ElementBase
	Handle native.Handle
}

// NativeMenuItem implements [goui.NativeMenuItemElement.NativeMenuItem].
func (e *nativeItemElement) NativeMenuItem() native.Handle {
	return e.Handle
}

// Destroy implements [goui.Element.Destroy].
func (e *nativeItemElement) Destroy() (err error) {
	return native.DestroyMenuItem(e.Handle)
}

// SetWidget implements [goui.Element.SetWidget].
func (e *nativeItemElement) SetWidget(ctx *goui.Context, w goui.WidgetBase) (err error) {
	oldItem, _ := e.Widget().(*Item)
	newItem := w.(*Item)
	if err = e.ElementBase.SetWidget(ctx, w); err != nil {
		return
	}

	if oldItem != nil {
		if oldItem.Title != newItem.Title {
			if err = native.SetMenuItemTitle(e.Handle, newItem.Title); err != nil {
				return
			}
		}
		if oldItem.Disabled != newItem.Disabled {
			if err = native.SetMenuItemDisabled(e.Handle, newItem.Disabled); err != nil {
				return
			}
		}
		return
	}

	if err = native.SetMenuItemTitle(e.Handle, newItem.Title); err != nil {
		return
	}
	if err = native.SetMenuItemDisabled(e.Handle, newItem.Disabled); err != nil {
		return
	}

	if newItem.OnSelect == nil {
		native.SetMenuItemOnClickListener(e.Handle, nil)
	} else {
		native.SetMenuItemOnClickListener(e.Handle, func() {
			newItem.OnSelect(ctx)
		})
	}
	return
}

// Separator is a menu separator widget that can be added to a [Menu].
type Separator struct {
	ID goui.ID
}

// WidgetID implements [goui.Widget.WidgetID].
func (sep *Separator) WidgetID() goui.ID {
	return sep.ID
}

// CreateElement implements [goui.Widget.CreateElement].
func (sep *Separator) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	handle, err := createItem(parent, "", true)
	if err != nil {
		return
	}
	element = &separatorElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}
	return
}

func (*Separator) ExclusiveWidgetMenu(goui.MenuItem) { /*Nop*/ }

type separatorElement struct {
	goui.ElementBase
	Handle native.Handle
}

// NativeMenuItem implements [goui.NativeMenuItemElement.NativeMenuItem].
func (e *separatorElement) NativeMenuItem() native.Handle {
	return e.Handle
}

// Destroy implements [goui.Element.Destroy].
func (e *separatorElement) Destroy() (err error) {
	return native.DestroyMenuItem(e.Handle)
}

// menuLayouter is a simple layouter for menu elements.
// Menu and menuitem do not need layouters, but we can't prevent users from
// adding non-menu(item) children to menu and menu item elements,
// so we provide menu and menuitem this type of layouter.
// Of course the non-menu(item) children will not be displayed in the actual menu,
// but in the ancestor native control(if any) or the root window.
type menuLayouter struct {
	goui.LayouterHelper
}

// Layout implements [goui.Layouter.Layout].
func (l *menuLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (metrics.Size, error) {
	return constraints.MinSize(), nil
}

// PositionAt implements [goui.Layouter.PositionAt].
func (l *menuLayouter) PositionAt(pt metrics.Point) (err error) {
	// Nop positioning.
	return nil
}

// Replayer implements [goui.Layouter.Replayer].
func (l *menuLayouter) Replayer() func(*goui.Context) error {
	return func(ctx *goui.Context) error {
		// Nop replaying to prevent the layout updating from going up to the window.
		return nil
	}
}

type StatefulMenu interface {
	goui.StatefulWidgetBase
	ExclusiveWidgetMenu(goui.Menu)
}

type statefulMenu struct {
	goui.StatefulHelper
}

func (*statefulMenu) ExclusiveWidgetMenu(goui.Menu) { /*Nop*/ }

// MenuState is the state associated with a [StatefulItem].
// See [goui.State] for more details.
type MenuState interface {
	Build() goui.Menu
	Destroy()
	Update(updater func()) error
}

type menuState struct {
	stateBase
	build func() goui.Menu // Can't be nil
}

func (s *menuState) Build() goui.Menu {
	return s.build()
}

func NewMenuState(ctx *goui.StateContext, build func() goui.Menu, destroy func()) MenuState {
	return &menuState{
		stateBase: stateBase{
			StateUpdater: goui.NewStateUpdater(ctx),
			destroy:      destroy,
		},
		build: build,
	}
}

// menuStateAdapter adapts a MenuState to a goui.State.
type menuStateAdapter struct {
	MenuState
}

func (a menuStateAdapter) Build() goui.WidgetBase {
	return a.MenuState.Build()
}

// NewStatefulMenu creates a new StatefulMenu with the given ID and state creator function.
func NewStatefulMenu(ID goui.ID, stateCreator func(ctx *goui.StateContext) MenuState) StatefulMenu {
	return &statefulMenu{
		StatefulHelper: goui.StatefulHelper{
			ID:           ID,
			StateCreator: func(ctx *goui.StateContext) goui.StateBase { return &menuStateAdapter{stateCreator(ctx)} },
		},
	}
}

type StatefulItem interface {
	goui.StatefulWidgetBase
	ExclusiveWidgetMenu(goui.MenuItem)
}

type statefulItem struct {
	goui.StatefulHelper
}

func (*statefulItem) ExclusiveWidgetMenu(goui.MenuItem) { /*Nop*/ }

// ItemState is the state associated with a [StatefulItem].
// See [goui.State] for more details.
type ItemState interface {
	Build() goui.MenuItem
	Destroy()
	Update(updater func()) error
}

// stateBase is the common base for [menuState] and [itemState].
type stateBase struct {
	goui.StateUpdater
	destroy func() // Can be nil
}

func (s *stateBase) Destroy() {
	if s.destroy != nil {
		s.destroy()
	}
}

// itemState is an implementation of ItemState.
type itemState struct {
	stateBase
	build func() goui.MenuItem // Can't be nil
}

func (s *itemState) Build() goui.MenuItem {
	return s.build()
}

func NewItemState(ctx *goui.StateContext, build func() goui.MenuItem, destroy func()) ItemState {
	return &itemState{
		stateBase: stateBase{
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

func (a itemStateAdapter) Build() goui.WidgetBase {
	return a.ItemState.Build()
}

// NewStatefulMenuItem creates a new StatefulMenuItem with the given ID and state creator function.
func NewStatefulItem(ID goui.ID, stateCreator func(ctx *goui.StateContext) ItemState) StatefulItem {
	return &statefulItem{
		StatefulHelper: goui.StatefulHelper{
			ID:           ID,
			StateCreator: func(ctx *goui.StateContext) goui.StateBase { return &itemStateAdapter{stateCreator(ctx)} },
		},
	}
}

type StatelessMenu interface {
	goui.StatelessWidgetBase
	ExclusiveWidgetMenu(goui.Menu)
}

type statelessMenu struct {
	goui.StatelessHelper
}

func (*statelessMenu) ExclusiveWidgetMenu(goui.Menu) { /*Nop*/ }

func NewStatelessMenu(ID goui.ID, builder func(ctx *goui.Context) goui.Menu) StatelessMenu {
	return &statelessMenu{
		StatelessHelper: goui.StatelessHelper{
			ID:      ID,
			Builder: func(ctx *goui.Context) goui.WidgetBase { return builder(ctx) },
		},
	}
}

type StatelessItem interface {
	goui.StatelessWidgetBase
	ExclusiveWidgetMenu(goui.MenuItem)
}

type statelessItem struct {
	goui.StatelessHelper
}

func (*statelessItem) ExclusiveWidgetMenu(goui.MenuItem) { /*Nop*/ }

func NewStatelessItem(ID goui.ID, builder func(ctx *goui.Context) goui.MenuItem) StatelessItem {
	return &statelessItem{
		StatelessHelper: goui.StatelessHelper{
			ID:      ID,
			Builder: func(ctx *goui.Context) goui.WidgetBase { return builder(ctx) },
		},
	}
}
