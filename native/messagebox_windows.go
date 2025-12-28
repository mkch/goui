package native

import (
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/gw/win32"
	"github.com/mkch/gw/win32/win32util"
)

type MessageBoxIcon int

const (
	MessageBoxIconNone MessageBoxIcon = iota
	MessageBoxIconInfo
	MessageBoxIconWarning
	MessageBoxIconQuestion
	MessageBoxIconError
)

type MessageBoxButton int

const (
	MessageBoxButtonOK MessageBoxButton = iota
	MessageBoxButtonYesNo
)

type MessageBoxReturn int

const (
	MessageBoxReturnCancel MessageBoxReturn = iota
	MessageBoxReturnOK
	MessageBoxReturnYes
	MessageBoxReturnNo
)

func MessageBox(parent Handle, title, message string, icon MessageBoxIcon, button MessageBoxButton) (ret MessageBoxReturn, err error) {
	var nativeParent win32.HWND
	if parent != nil {
		nativeParent = parent.(winBase).HWND()
	}
	var nativeType win32.MESSAGE_BOX_TYPE
	switch button {
	case MessageBoxButtonYesNo:
		nativeType |= win32.MB_YESNO
	}
	switch icon {
	case MessageBoxIconInfo:
		nativeType |= win32.MB_ICONINFORMATION
	case MessageBoxIconWarning:
		nativeType |= win32.MB_ICONWARNING
	case MessageBoxIconQuestion:
		nativeType |= win32.MB_ICONQUESTION
	case MessageBoxIconError:
		nativeType |= win32.MB_ICONERROR
	}
	id, err := win32util.MessageBox(nativeParent, message, title, nativeType)
	if err != nil {
		return
	}
	err = errortrace.WithStack(err)
	switch id {
	case win32.IDOK:
		ret = MessageBoxReturnOK
	case win32.IDYES:
		ret = MessageBoxReturnYes
	case win32.IDNO:
		ret = MessageBoxReturnNo
	case win32.IDCANCEL:
		ret = MessageBoxReturnCancel
	}
	return
}
