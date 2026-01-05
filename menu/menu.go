// Package menu implements menu widgets.
package menu

import (
	"errors"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// Menu is a menu widget that can contain items of type [Item] and [Separator].
type Menu struct {
	ID goui.ID
	// Items are the menu's child widgets.
	// Only widgets representing menu items should be added; others will not display correctly.
	Items []goui.Widget
}

// WidgetID implements [goui.Widget.WidgetID].
func (m *Menu) WidgetID() goui.ID {
	return m.ID
}

// CreateElement implements [goui.Widget.CreateElement].
func (m *Menu) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	return createMenuElement(ctx, parent, true)
}

// createMenuElement is a helper function to create a popup or window menu element.
func createMenuElement(ctx *goui.Context, parent goui.Element, popup bool) (element goui.Element, err error) {
	handle := native.CreateMenu(popup)
	if opener := goui.NativeMenuItem(ctx, parent); opener != nil {
		if err = native.SetMenuItemSubmenu(opener, handle); err != nil {
			return
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
func (m *Menu) Child(index int) goui.Widget {
	return m.Items[index]
}

// Exclusive implements [goui.Container.Exclusive].
func (m *Menu) Exclusive(goui.Container) { /*Nop*/ }

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
	return createMenuElement(ctx, parent, false)
}

// NumChildren implements [goui.Container.NumChildren].
func (m *WindowMenu) NumChildren() int {
	return len(m.Items)
}

// Child implements [goui.Container.Child].
func (m *WindowMenu) Child(index int) goui.Widget {
	return m.Items[index]
}

// Exclusive implements [goui.Container.Exclusive].
func (m *WindowMenu) Exclusive(goui.Container) { /*Nop*/ }

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
	// Only widgets representing menus should be assigned; others will not display correctly.
	Submenu  goui.Widget
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
func (item *Item) Child(index int) goui.Widget {
	if item.Submenu == nil || index != 0 {
		panic("index out of range")
	}
	return item.Submenu
}

// Exclusive implements [goui.Container.Exclusive].
func (item *Item) Exclusive(goui.Container) { /*Nop*/ }

// ErrNoParentMenu is returned by [Item.CreateElement] when trying to create a MenuItem element without a parent menu element.
var ErrNoParentMenu = errors.New("cannot use menu items out of a menu")

// CreateElement implements [goui.Widget.CreateElement].
func (item *Item) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	nativeParent := goui.LookupNativeMenuParent(ctx, parent)
	if nativeParent == nil {
		return nil, ErrNoParentMenu
	}
	handle, err := native.CreateMenuItem(nativeParent, item.Title, false)
	if err != nil {
		return nil, err
	}
	return &nativeItemElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}, nil
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
func (e *nativeItemElement) SetWidget(ctx *goui.Context, w goui.Widget) (err error) {
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
func (sep *Separator) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	nativeParent := goui.LookupNativeMenuParent(ctx, parent)
	if nativeParent == nil {
		return nil, ErrNoParentMenu
	}
	handle, err := native.CreateMenuItem(nativeParent, "", true)
	if err != nil {
		return nil, err
	}
	return &separatorElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}, nil
}

type separatorElement struct {
	goui.ElementBase
	Handle native.Handle
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
func (l *menuLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (goui.Size, error) {
	return constraints.MinSize(), nil
}

// PositionAt implements [goui.Layouter.PositionAt].
func (l *menuLayouter) PositionAt(x, y metrics.DP) (err error) {
	for child := range l.Children() {
		if err = child.PositionAt(x, y); err != nil {
			return
		}
	}
	return
}

// Replayer implements [goui.Layouter.Replayer].
func (l *menuLayouter) Replayer() func(*goui.Context) error {
	return func(ctx *goui.Context) error {
		// Nop replaying to prevent the layout updating from going up to the window.
		return nil
	}
}
