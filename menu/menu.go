// Package menu implements menu widgets.
package menu

import (
	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// Menu is a menu widget that can contain menu items.
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
	return createMenuElement(ctx, parent, true)
}

// NumChildren implements [goui.Container.NumChildren].
func (m *Menu) NumChildren() int {
	return len(m.Items)
}

// Child implements [goui.Container.Child].
func (m *Menu) Child(index int) goui.AbstractWidget {
	return m.Items[index]
}

func (*Menu) ExclusiveType(marker.TypeMenu)      { /*Nop*/ }
func (*Menu) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

// nativeItemParent searches the element and its ancestors for the nearest [goui.NativeMenuItemElement]
// and returns its handle, or nil if not found.
// The search stops and returns [ErrWrongParent] error when encountering an element is of type
// [goui.NativeMenuItemElement]  and whose widget is neither a [StatelessMenu] or [StatefulMenu].
func nativeItemParent(element goui.Element) (native.Handle, error) {
	item, err := nativeParent[goui.NativeMenuItemElement, StatelessMenu, StatefulMenu](element)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return item.NativeMenuItem(), nil
}

// nativeMenuParent searches the element and its ancestors for the nearest [goui.NativeMenuElement]
// and returns its handle, or nil if not found.
// The search stops and returns [ErrWrongParent] error when encountering an element is of type
// [goui.NativeMenuElement]  and whose widget is neither a [StatelessItem] or [StatefulItem].
func nativeMenuParent(element goui.Element) (native.Handle, error) {
	menu, err := nativeParent[goui.NativeMenuElement, StatelessItem, StatefulItem](element)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, nil
	}
	return menu.NativeMenu(), nil
}

// nativeParent is a helper function for [nativeMenuParent] and [nativeItemParent].
// It searches the element and its ancestors for the nearest element of type ElementType.
// and returns it via ret, or nil if not found.
// The search stops and returns goui.ErrWrongParent error when encountering an element whose type
// is not T and whose widget is neither a [StatelessType] or [StatefulType].
func nativeParent[
	ElementType goui.Element,
	StatelessType goui.AbstractStatelessWidget,
	StatefulType goui.AbstractStatefulWidget,
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
func createMenuElement(ctx *goui.Context, parent goui.Element, popup bool) (element goui.Element, err error) {
	handle, err := goui.OS().NewMenu(popup)
	if err != nil {
		return
	}
	if parent != nil {
		// If parent is given, it must represent a opener
		// menu item to which this menu is a submenu.
		var opener native.Handle
		if opener, err = nativeItemParent(parent); err != nil {
			err = errortrace.ErrorfStack("create menu failed: %w", err)
			return
		}
		if opener != nil {
			if err = goui.OS().MenuItem_SetSubmenu(opener, handle); err != nil {
				return
			}
		}
	}
	return &nativeMenuElement{
		ElementHelper: goui.ElementHelper{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}, nil
}

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
func (m *WindowMenu) Child(index int) goui.AbstractWidget {
	return m.Items[index]
}

func (*WindowMenu) ExclusiveType(marker.TypeMenu)      { /*Nop*/ }
func (*WindowMenu) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

// nativeMenuElement is an implementation of [goui.NativeMenuElement]
// that represents a native menu.
type nativeMenuElement struct {
	goui.ElementHelper
	Handle native.Handle
}

// NativeMenu implements [goui.NativeMenuElement.NativeMenu].
func (e *nativeMenuElement) NativeMenu() native.Handle {
	return e.Handle
}

// Destroy implements [goui.Element.Destroy].
func (e *nativeMenuElement) Destroy(ctx *goui.Context) (err error) {
	return goui.OS().Menu_Destroy(e.Handle)
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
func (item *Item) Child(index int) goui.AbstractWidget {
	if item.Submenu == nil || index != 0 {
		panic("index out of range")
	}
	return item.Submenu
}

func (*Item) ExclusiveType(marker.TypeMenuItem)  { /*Nop*/ }
func (*Item) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

// CreateElement implements [goui.Widget.CreateElement].
func (item *Item) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	handle, err := createItem(ctx, parent, item.Title, false)
	if err != nil {
		return
	}
	element = &nativeItemElement{
		ElementHelper: goui.ElementHelper{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}
	return
}

// createItem is a helper function to create a menu item under the given parent element.
// If separator is true, a separator item is created.
func createItem(ctx *goui.Context, parent goui.Element, title string, separator bool) (handle native.Handle, err error) {
	parentMenu, err := nativeMenuParent(parent)
	if err == nil && parentMenu == nil {
		// Menu item must have a parent menu
		err = goui.ErrWrongParent
	}
	if err != nil {
		err = errortrace.ErrorfStack("create menu %s failed: %w", gg.If(separator, "separator", "item"), err)
		return
	}
	handle, err = goui.OS().NewMenuItem(parentMenu, title, separator)
	return
}

// nativeItemElement is an implementation of [goui.NativeMenuItemElement]
// that represents a native menu item.
type nativeItemElement struct {
	goui.ElementHelper
	Handle native.Handle
}

// NativeMenuItem implements [goui.NativeMenuItemElement.NativeMenuItem].
func (e *nativeItemElement) NativeMenuItem() native.Handle {
	return e.Handle
}

// Destroy implements [goui.Element.Destroy].
func (e *nativeItemElement) Destroy(ctx *goui.Context) (err error) {
	return goui.OS().MenuItem_Destroy(e.Handle)
}

// SetWidget implements [goui.Element.SetWidget].
func (e *nativeItemElement) SetWidget(ctx *goui.Context, w goui.AbstractWidget) (err error) {
	oldItem, _ := e.Widget().(*Item)
	newItem := w.(*Item)
	if err = e.ElementHelper.SetWidget(ctx, w); err != nil {
		return
	}

	if oldItem != nil {
		if oldItem.Title != newItem.Title {
			if err = goui.OS().MenuItem_SetTitle(e.Handle, newItem.Title); err != nil {
				return
			}
		}
		if oldItem.Disabled != newItem.Disabled {
			if err = goui.OS().MenuItem_SetDisabled(e.Handle, newItem.Disabled); err != nil {
				return
			}
		}
		return
	}

	if err = goui.OS().MenuItem_SetTitle(e.Handle, newItem.Title); err != nil {
		return
	}
	if err = goui.OS().MenuItem_SetDisabled(e.Handle, newItem.Disabled); err != nil {
		return
	}

	if newItem.OnSelect == nil {
		goui.OS().MenuItem_SetOnClickListener(e.Handle, nil)
	} else {
		goui.OS().MenuItem_SetOnClickListener(e.Handle, func() {
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
	handle, err := createItem(ctx, parent, "", true)
	if err != nil {
		return
	}
	element = &separatorElement{
		ElementHelper: goui.ElementHelper{
			ElementLayouter: &menuLayouter{},
		},
		Handle: handle,
	}
	return
}

func (*Separator) ExclusiveType(marker.TypeMenuItem) { /*Nop*/ }

type separatorElement struct {
	goui.ElementHelper
	Handle native.Handle
}

// NativeMenuItem implements [goui.NativeMenuItemElement.NativeMenuItem].
func (e *separatorElement) NativeMenuItem() native.Handle {
	return e.Handle
}

// Destroy implements [goui.Element.Destroy].
func (e *separatorElement) Destroy(ctx *goui.Context) (err error) {
	return goui.OS().MenuItem_Destroy(e.Handle)
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
func (l *menuLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
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

// Items is a menu that contains menu items.
// It is convenient to use when defining menus inline.
// See example of [Menu] for usage.
type Items []goui.MenuItem

// WidgetID implements [goui.Menu.WidgetID] and always returns nil.
func (items Items) WidgetID() goui.ID {
	return nil
}

// CreateElement implements [goui.Menu.CreateElement].
func (items Items) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return createMenuElement(ctx, parent, true)
}

// NumChildren implements [goui.AbstractContainer.NumChildren].
func (items Items) NumChildren() int {
	return len(items)
}

// Child implements [goui.AbstractContainer.Child].
func (items Items) Child(index int) goui.AbstractWidget {
	return items[index]
}

func (Items) ExclusiveType(marker.TypeMenu)      { /*Nop*/ }
func (Items) ExclusiveKind(marker.KindContainer) { /*Nop*/ }
