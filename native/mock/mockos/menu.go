package mockos

import "errors"

var menus = make(map[Handle]*menu)
var menuItems = make(map[Handle]*menuItem)

var ErrInvalidMenu = errors.New("invalid menu or menu item handle")

type menu []Handle

type menuItem struct {
	separator bool
	onClick   func()
	title     string
	disabled  bool
	subMenu   Handle
}

func NewMenu() Handle {
	var m menu
	handle := newHandle()
	menus[handle] = &m
	return handle
}

func DestroyMenu(h Handle) error {
	m, ok := menus[h]
	if !ok {
		return ErrInvalidMenu
	}
	for _, item := range *m {
		if err := DestroyMenuItem(item); err != nil {
			return err
		}
	}

	delete(menus, h)
	return nil
}

func DestroyMenuItem(h Handle) error {
	item, ok := menuItems[h]
	if !ok {
		return ErrInvalidMenu
	}
	if item.subMenu != nil {
		if err := DestroyMenu(item.subMenu); err != nil {
			return err
		}
	}
	delete(menuItems, h)
	return nil
}

func NewMenuItem(parent Handle, title string, separator bool) (Handle, error) {
	subMenu, ok := menus[parent]
	if !ok {
		return nil, ErrInvalidMenu
	}
	handle := newHandle()
	menuItems[handle] = &menuItem{
		separator: separator,
		title:     title,
	}
	*subMenu = append(*subMenu, handle)
	return handle, nil
}

func SetMenuItemTitle(item Handle, title string) error {
	menuItem, ok := menuItems[item]
	if !ok {
		return ErrInvalidMenu
	}
	menuItem.title = title
	return nil
}

func SetMenuItemDisabled(item Handle, disabled bool) error {
	menuItem, ok := menuItems[item]
	if !ok {
		return ErrInvalidMenu
	}
	menuItem.disabled = disabled
	return nil
}

func SetMenuItemSubmenu(item Handle, submenu Handle) error {
	menuItem, ok := menuItems[item]
	if !ok {
		return ErrInvalidMenu
	}
	if _, ok = menus[submenu]; !ok {
		return ErrInvalidMenu
	}
	if oldSubMenu := menuItem.subMenu; oldSubMenu != nil {
		if err := DestroyMenu(oldSubMenu); err != nil {
			return err
		}
	}
	menuItem.subMenu = submenu
	return nil
}

func SetMenuItemOnClickListener(item Handle, listener func()) error {
	menuItem, ok := menuItems[item]
	if !ok {
		return ErrInvalidMenu
	}
	menuItem.onClick = listener
	return nil
}
