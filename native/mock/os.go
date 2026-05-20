package mock

import (
	"image/color"
	"iter"

	"github.com/mkch/gg/errorcheck"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
	"github.com/mkch/goui/native/mock/mockos"
)

func must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	return errorcheck.Must2(errortrace.Panic, v1, v2, err)
}

type OS struct {
	// listener -> window
	mouseEventListeners map[*native.MouseEventListener]native.Handle

	lastMessageBoxParams MessageBoxParams
	nextMessageBoxReturn struct {
		ret native.MessageBoxReturn
		err error
	}
}

func NewOS() *OS {
	return &OS{}
}
func (*OS) App_Run(f func()) int {
	return mockos.Run(f)
}
func (*OS) App_Post(f func()) error {
	return mockos.Post(f)
}
func (*OS) App_Quit(exitCode int) error {
	return mockos.PostQuitMessage(exitCode)
}

func callMouseEventListeners(win mockos.Handle, x, y metrics.DP,
	listeners map[*native.MouseEventListener]native.Handle,
	eventFunc func(listener native.MouseEventListener, parent native.Handle, x, y metrics.DP)) {
	var screenX, screenY = must2(mockos.ClientToScreen(win, x, y))
	for listener, parent := range listeners {
		clientX, clientY := must2(mockos.ScreenToClient(parent.(mockos.Handle), screenX, screenY))
		eventFunc(*listener, parent, clientX, clientY)
	}
}

func (os *OS) App_AddMouseEventListener(win native.Handle, listener native.MouseEventListener) (remove func()) {
	type privateKey int
	const listenersKey privateKey = 0

	if os.mouseEventListeners == nil {
		os.mouseEventListeners = make(map[*native.MouseEventListener]native.Handle)
	}
	mockos.SetMessageDispatcher(func(msg *mockos.Msg, prev func(*mockos.Msg) any) any {
		switch evt := msg.Message.(type) {
		case mockos.MsgMouseLeftDown:
			callMouseEventListeners(msg.Window, evt.X, evt.Y, os.mouseEventListeners,
				native.MouseEventListener.OnMousePrimaryDown)
		case mockos.MsgMouseLeftUp:
			callMouseEventListeners(msg.Window, evt.X, evt.Y, os.mouseEventListeners,
				native.MouseEventListener.OnMousePrimaryUp)
		case mockos.MsgMouseRightDown:
			callMouseEventListeners(msg.Window, evt.X, evt.Y, os.mouseEventListeners,
				native.MouseEventListener.OnMouseSecondaryDown)
		case mockos.MsgMouseRightUp:
			callMouseEventListeners(msg.Window, evt.X, evt.Y, os.mouseEventListeners,
				native.MouseEventListener.OnMouseSecondaryUp)
		case mockos.MsgMouseMiddleDown:
			callMouseEventListeners(msg.Window, evt.X, evt.Y, os.mouseEventListeners,
				native.MouseEventListener.OnMouseMiddleDown)
		case mockos.MsgMouseMiddleUp:
			callMouseEventListeners(msg.Window, evt.X, evt.Y, os.mouseEventListeners,
				native.MouseEventListener.OnMouseMiddleUp)
		}
		return prev(msg)
	})

	os.mouseEventListeners[&listener] = win
	return func() {
		delete(os.mouseEventListeners, &listener)
	}
}

func (*OS) NewWindow(title string, size metrics.Size) (h native.Handle, err error) {
	return mockos.NewWindow(title, size)
}

func (*OS) Window_Close(win native.Handle) error {
	return mockos.CloseWindow(win.(mockos.Handle))
}
func (*OS) Window_Invalidate(win native.Handle, rect *metrics.Rect) error {
	return mockos.InvalidateWindow(win.(mockos.Handle), rect)
}
func (*OS) Window_Destroy(win native.Handle) error {
	return mockos.DestroyWindow(win.(mockos.Handle))
}
func (*OS) Window_SetOnSizeChangedListener(win native.Handle, listener func(size metrics.Size)) error {
	return mockos.SetOnSizeChangedListener(win.(mockos.Handle), listener)
}
func (*OS) Window_SetOnDestroyListener(win native.Handle, listener func()) error {
	return mockos.SetOnDestroyListener(win.(mockos.Handle), listener)
}
func (*OS) Window_SetOnCloseListener(win native.Handle, listener func() bool) error {
	return mockos.SetOnCloseListener(win.(mockos.Handle), listener)
}
func (*OS) Window_ClientRect(win native.Handle) (rect metrics.Rect, err error) {
	return mockos.ClientRect(win.(mockos.Handle))
}
func (*OS) Window_Menu(win native.Handle) (native.Handle, error) {
	return mockos.WindowMenu(win.(mockos.Handle))
}
func (*OS) Window_SetMenu(win native.Handle, m native.Handle) error {
	return mockos.SetMenu(win.(mockos.Handle), m.(mockos.Handle))
}
func (*OS) Window_RefreshMenu(win native.Handle) error {
	return mockos.RefreshMenu(win.(mockos.Handle))
}
func (*OS) Window_TrackPopupMenu(win native.Handle, menuToTrack native.Handle, spec *native.TrackPopupSpec) error {
	return mockos.TrackPopupMenu(win.(mockos.Handle), menuToTrack.(mockos.Handle), spec)
}
func (*OS) Window_EnableDrawDebugRect(win native.Handle, rects func() iter.Seq[native.DebugRect]) (layer native.Handle, err error) {
	return mockos.EnableDrawDebugRect(win.(mockos.Handle), rects)
}

func (*OS) NewMenu(popup bool) (native.Handle, error) {
	return mockos.NewMenu(), nil
}
func (*OS) Menu_Destroy(m native.Handle) error {
	return mockos.DestroyMenu(m.(mockos.Handle))
}

func (*OS) NewMenuItem(parent native.Handle, title string, separator bool) (handle native.Handle, err error) {
	return mockos.NewMenuItem(parent.(mockos.Handle), title, separator)
}
func (*OS) MenuItem_Destroy(item native.Handle) error {
	return mockos.DestroyMenuItem(item.(mockos.Handle))
}
func (*OS) MenuItem_SetTitle(item native.Handle, title string) error {
	return mockos.SetMenuItemTitle(item.(mockos.Handle), title)
}
func (*OS) MenuItem_SetDisabled(item native.Handle, disabled bool) error {
	return mockos.SetMenuItemDisabled(item.(mockos.Handle), disabled)
}
func (*OS) MenuItem_SetSubmenu(item native.Handle, submenu native.Handle) error {
	return mockos.SetMenuItemSubmenu(item.(mockos.Handle), submenu.(mockos.Handle))
}
func (*OS) MenuItem_SetOnClickListener(item native.Handle, listener func()) {
	mockos.SetMenuItemOnClickListener(item.(mockos.Handle), listener)
}

// Debug_MenuItems returns the menu items of the given parent menu.
// For debug purposes only.
func (*OS) Debug_MenuItems(parent native.Handle) ([]native.Handle, error) {
	handles, err := mockos.MenuItems(parent.(mockos.Handle))
	if err != nil {
		return nil, err
	}
	nativeHandles := make([]native.Handle, len(handles))
	for i, h := range handles {
		nativeHandles[i] = h
	}
	return nativeHandles, nil
}

// Debug_MenuItemTitle returns the title of the given menu item.
// For debug purposes only.
func (*OS) Debug_MenuItemTitle(item native.Handle) (string, error) {
	return mockos.MenuItemTitle(item.(mockos.Handle))
}

// Debug_MenuItemIsSeparator returns whether the given menu item is a separator.
func (*OS) Debug_MenuItemIsSeparator(item native.Handle) (bool, error) {
	return mockos.MenuItemIsSeparator(item.(mockos.Handle))
}

// Debug_MenuItemSubMenu returns the submenu of the given menu item, or nil if it has no submenu.
func (*OS) Debug_MenuItemSubMenu(item native.Handle) (native.Handle, error) {
	return mockos.MenuItemSubMenu(item.(mockos.Handle))
}

func (*OS) NewButton(parent native.Handle, title string) (handle native.Handle, err error) {
	return mockos.NewButton(parent.(mockos.Handle), title)
}
func (*OS) Button_SetOnClickListener(btn native.Handle, onClick func()) error {
	return mockos.SetButtonOnClickListener(btn.(mockos.Handle), onClick)
}
func (*OS) Button_SetLabel(btn native.Handle, label string) error {
	return mockos.SetText(btn.(mockos.Handle), label)
}
func (*OS) Button_MinimumSize(btn native.Handle, label string) (size metrics.Size, err error) {
	return mockos.ButtonMinimumSize(btn.(mockos.Handle), label)
}

func (*OS) NewLabel(parent native.Handle, title string) (handle native.Handle, err error) {
	return mockos.NewLabel(parent.(mockos.Handle), title)
}
func (*OS) Label_SetMultiline(lbl native.Handle, multiline bool) error {
	return mockos.SetLabelMultiline(lbl.(mockos.Handle), multiline)
}
func (*OS) Label_SetTextAlignment(lbl native.Handle, alignment native.TextAlignment) error {
	return mockos.SetLabelTextAlignment(lbl.(mockos.Handle), alignment)
}
func (*OS) Label_SetText(handle native.Handle, text string) error {
	return mockos.SetText(handle.(mockos.Handle), text)
}
func (*OS) Label_SetBackgroundColor(lbl native.Handle, color *color.NRGBA) error {
	return mockos.SetLabelBackgroundColor(lbl.(mockos.Handle), color)
}

func (*OS) NewTextField(parent native.Handle, initialValue string, password bool) (handle native.Handle, err error) {
	return mockos.NewTextField(parent.(mockos.Handle), initialValue, password)
}
func (*OS) TextField_Text(tf native.Handle) (text string, err error) {
	return mockos.Text(tf.(mockos.Handle))
}
func (*OS) TextField_SetText(tf native.Handle, text string) error {
	return mockos.SetText(tf.(mockos.Handle), text)
}

func (*OS) Control_SetDimensions(ctrl native.Handle, rect metrics.Rect) error {
	return mockos.SetControlDimensions(ctrl.(mockos.Handle), rect)
}
func (*OS) Control_SetEnabled(ctrl native.Handle, enabled bool) error {
	return mockos.SetControlEnabled(ctrl.(mockos.Handle), enabled)
}
func (*OS) Control_Enabled(ctrl native.Handle) (bool, error) {
	return mockos.ControlEnabled(ctrl.(mockos.Handle))
}
func (*OS) Control_SetOnDestroyListener(handle native.Handle, listener func()) error {
	return mockos.SetOnDestroyListener(handle.(mockos.Handle), listener)
}
func (*OS) Control_Destroy(win native.Handle) error {
	return mockos.DestroyWindow(win.(mockos.Handle))
}
func (*OS) Control_TextDrawingSize(ctrl native.Handle, text string, multiline bool, maxWidth metrics.DP) (size metrics.Size, err error) {
	return mockos.TextDrawingSize(ctrl.(mockos.Handle), text, multiline, maxWidth)
}

func (*OS) NewPanel(parent native.Handle) (handle native.Handle, err error) {
	return mockos.NewPanel(parent.(mockos.Handle))
}
func (*OS) Panel_SetBackgroundColor(handle native.Handle, color *color.NRGBA) error {
	return mockos.SetPanelBackgroundColor(handle.(mockos.Handle), color)
}

func (*OS) Util_ClientCoordinatesConv(from, to native.Handle, x, y metrics.DP) (newX, newY metrics.DP, err error) {
	screenX, screenY, err := mockos.ClientToScreen(from.(mockos.Handle), x, y)
	if err != nil {
		return 0, 0, err
	}
	return mockos.ScreenToClient(to.(mockos.Handle), screenX, screenY)
}

func (*OS) Util_ClientToScreen(win native.Handle, x, y metrics.DP) (screenX, screenY metrics.DP, err error) {
	return mockos.ClientToScreen(win.(mockos.Handle), x, y)
}

type MessageBoxParams struct {
	Parent         native.Handle
	Title, Message string
	Icon           native.MessageBoxIcon
	Button         native.MessageBoxButton
}

func (os *OS) MessageBox(parent native.Handle, title, message string, icon native.MessageBoxIcon, button native.MessageBoxButton) (ret native.MessageBoxReturn, err error) {
	os.lastMessageBoxParams = MessageBoxParams{
		Parent:  parent,
		Title:   title,
		Message: message,
		Icon:    icon,
		Button:  button,
	}
	return os.nextMessageBoxReturn.ret, os.nextMessageBoxReturn.err
}

func (os *OS) Debug_LastMessageBoxParams() MessageBoxParams {
	return os.lastMessageBoxParams
}

func (os *OS) Debug_SetNextMessageBoxReturn(ret native.MessageBoxReturn, err error) {
	os.nextMessageBoxReturn.ret = ret
	os.nextMessageBoxReturn.err = err
}
