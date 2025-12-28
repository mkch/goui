package messagebox

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/native"
)

type Icon native.MessageBoxIcon

const (
	IconNone     = Icon(native.MessageBoxIconNone)
	IconInfo     = Icon(native.MessageBoxIconInfo)
	IconWarning  = Icon(native.MessageBoxIconWarning)
	IconQuestion = Icon(native.MessageBoxIconQuestion)
	IconError    = Icon(native.MessageBoxIconError)
)

type Button native.MessageBoxButton

const (
	ButtonOK    = Button(native.MessageBoxButtonOK)
	ButtonYesNo = Button(native.MessageBoxButtonYesNo)
)

type Return native.MessageBoxReturn

const (
	ReturnCancel = Return(native.MessageBoxReturnCancel)
	ReturnOK     = Return(native.MessageBoxReturnOK)
	ReturnYes    = Return(native.MessageBoxReturnYes)
	ReturnNo     = Return(native.MessageBoxReturnNo)
)

// MessageBox shows a message box with the given title, message, icon and buttons
// If ctx is non-nil, the message box is associated with the context's window.
func Show(ctx *goui.Context, title, message string, icon Icon, button Button) (ret Return, err error) {
	var parent native.Handle
	if ctx != nil {
		parent = ctx.NativeWindow()
	}
	id, err := native.MessageBox(parent, title, message, native.MessageBoxIcon(icon), native.MessageBoxButton(button))
	ret = Return(id)
	return
}
