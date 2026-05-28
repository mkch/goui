package mockos

import (
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type topLevelWindow struct {
	abstractWindow
	onClose func() bool
	menu    Handle
}

var defaultWindowLeftTop metrics.Point

func NewTopLevelWindow(title string, rect *metrics.Rect) *topLevelWindow {
	var win = &topLevelWindow{}
	initAbstractWindow(&win.abstractWindow, rect)
	win.SetWndProc(func(msg *Msg, prev func(*Msg) any) any {
		switch msg.Message.(type) {
		case MsgClosing:
			if win.onClose != nil {
				return win.onClose()
			}
		case MsgDestroyed:
			if win.menu != nil {
				DestroyMenu(win.menu)
			}
		}
		return prev(msg)
	})
	win.SetText(title)
	return win
}

func (win *topLevelWindow) SetOnCloseListener(onClose func() bool) {
	win.onClose = onClose
}

func (win *topLevelWindow) SetMenu(menu Handle) {
	if oldMenu := win.menu; oldMenu != nil {
		DestroyMenu(oldMenu)
	}
	win.menu = menu
}

func (win *topLevelWindow) Menu() Handle {
	return win.menu
}

func (win *topLevelWindow) RefreshMenu() error {
	return nil
}

func (win *topLevelWindow) TrackPopupMenu(menuToTrack Handle, spec *native.TrackPopupSpec) error {
	if menuToTrack == nil {
		return ErrInvalidMenu
	}
	return nil
}
