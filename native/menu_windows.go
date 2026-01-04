package native

import (
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/gw/menu"
	"github.com/mkch/gw/window"
)

func SetWindowMenu(win Handle, m Handle) (err error) {
	//nativeMenu.SetPopup(false) // Window menu should not be popup
	err = win.(*window.Window).SetMenu(m.(*menu.Menu))
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func CreateMenu(popup bool) Handle {
	return menu.New(popup)
}

func IsMenuPopup(m Handle) bool {
	return m.(*menu.Menu).Popup()
}

func DestroyMenu(m Handle) error {
	return m.(*menu.Menu).Destroy()
}

func CreateMenuItem(parent Handle, title string, separator bool) (handle Handle, err error) {
	handle, err = parent.(*menu.Menu).InsertItem(-1, &menu.ItemSpec{
		Title:     title,
		Separator: separator,
	})
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func DestroyMenuItem(item Handle) (err error) {
	nativeItem := item.(*menu.Item)
	err = nativeItem.Menu().DeleteItem(nativeItem)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func SetMenuItemTitle(item Handle, title string) (err error) {
	err = item.(*menu.Item).SetTitle(title)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func SetMenuItemDisabled(item Handle, disabled bool) (err error) {
	err = item.(*menu.Item).SetDisabled(disabled)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func SetMenuItemSubmenu(item Handle, submenu Handle) (err error) {
	err = item.(*menu.Item).SetSubmenu(submenu.(*menu.Menu))
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func SetMenuItemSeparator(item Handle, separator bool) (err error) {
	err = item.(*menu.Item).SetSeparator(separator)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func SetMenuItemOnClickListener(item Handle, listener func()) {
	item.(*menu.Item).OnClick = listener
}
