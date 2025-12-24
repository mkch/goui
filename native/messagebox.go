package native

import (
	"github.com/mkch/gw/win32"
	"github.com/mkch/gw/win32/win32util"
)

type MessageBoxIcon int

const (
	MessageBoxNone MessageBoxIcon = iota
	MessageBoxIconInfo
	MessageBoxIconWarning
	MessageBoxIconError
)

func MessageBox(parent Handle, title, message string, icon MessageBoxIcon) {
	var nativeParent win32.HWND
	if parent != nil {
		nativeParent = parent.(winBase).HWND()
	}
	var nativeType win32.MESSAGE_BOX_TYPE
	switch icon {
	case MessageBoxNone:
		nativeType = win32.MB_OK
	case MessageBoxIconInfo:
		nativeType = win32.MB_OK | win32.MB_ICONINFORMATION
	case MessageBoxIconWarning:
		nativeType = win32.MB_OK | win32.MB_ICONWARNING
	case MessageBoxIconError:
		nativeType = win32.MB_OK | win32.MB_ICONERROR
	}
	win32util.MessageBox(nativeParent, message, title, nativeType)
}
