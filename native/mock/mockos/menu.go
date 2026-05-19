package mockos

import (
	"errors"
	"slices"
)

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
	parentMenu, ok := menus[parent]
	if !ok {
		return nil, ErrInvalidMenu
	}
	handle := newHandle()
	menuItems[handle] = &menuItem{
		separator: separator,
		title:     title,
	}
	*parentMenu = append(*parentMenu, handle)
	return handle, nil
}

// MenuItems returns the menu items of the given menu handle.
func MenuItems(parent Handle) ([]Handle, error) {
	parentMenu, ok := menus[parent]
	if !ok {
		return nil, ErrInvalidMenu
	}
	return slices.Clone(*parentMenu), nil
}

// MenuItemIsSeparator returns whether the given menu item is a separator.
func MenuItemIsSeparator(item Handle) (bool, error) {
	menuItem, ok := menuItems[item]
	if !ok {
		return false, ErrInvalidMenu
	}
	return menuItem.separator, nil
}

// MenuItemSubMenu returns the submenu of the given menu item handle, or nil if it has no submenu.
func MenuItemSubMenu(item Handle) (Handle, error) {
	menuItem, ok := menuItems[item]
	if !ok {
		return nil, ErrInvalidMenu
	}
	return menuItem.subMenu, nil
}

func SetMenuItemTitle(item Handle, title string) error {
	menuItem, ok := menuItems[item]
	if !ok {
		return ErrInvalidMenu
	}
	menuItem.title = title
	return nil
}

// MenuItemTitle returns the title of the given menu item handle.
func MenuItemTitle(item Handle) (string, error) {
	menuItem, ok := menuItems[item]
	if !ok {
		return "", ErrInvalidMenu
	}
	return menuItem.title, nil
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
