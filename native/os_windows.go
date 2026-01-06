package native

import (
	"image/color"
	"unsafe"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/gw/app"
	"github.com/mkch/gw/app/gwapp"
	"github.com/mkch/gw/button"
	"github.com/mkch/gw/edit"
	"github.com/mkch/gw/menu"
	"github.com/mkch/gw/panel"
	"github.com/mkch/gw/static"
	"github.com/mkch/gw/win32"
	"github.com/mkch/gw/win32/win32util"
	"github.com/mkch/gw/window"
)

type App any

func NewApp() App {
	return gwapp.New()
}

func App_Run(app App) int {
	return app.(*gwapp.GwApp).Run()
}

func App_Post(app App, f func()) error {
	return app.(*gwapp.GwApp).Post(f)
}

func App_Quit(app App, exitCode int) {
	app.(*gwapp.GwApp).Quit(exitCode)
}

type MsgProc = gwapp.MessageDispatcher

// SetMessageDispatcher sets a dispatcher for windows message dispatching.
// The default message dispatcher is [win32.DispatchMessageW].
func SetMessageDispatcher(msgProc MsgProc) {
	app.SetMessageDispatcher(msgProc)
}

// Handle represents a platform-specific GUI object.
type Handle any

// CreateWindow creates a native window with the specified configuration.
func CreateWindow(title string, width, height metrics.DP) (handle Handle, err error) {
	win, err := window.New(&window.Spec{
		Text:  title,
		Style: win32.WS_OVERLAPPEDWINDOW | win32.WS_VISIBLE,
		X:     window.CW_USEDEFAULT,
	})
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	dpi, err := win32.GetDpiForWindow(win.HWND())
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	win32.SetWindowPos(win.HWND(), win32.HWND(0),
		0, 0,
		win32.INT(width.Px(uint(dpi))), win32.INT(height.Px(uint(dpi))),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE|win32.SWP_NOMOVE)
	win.Show(win32.SW_SHOWNORMAL)
	handle = win
	return
}

func CloseWindow(handle Handle) error {
	_, err := win32.SendMessageW(handle.(*window.Window).HWND(), win32.WM_CLOSE, 0, 0)
	return err
}

func InvalidateWindow(handle Handle) error {
	err := handle.(winBase).InvalidateRect(nil, true)
	return errortrace.WithStack(err)
}

type winBase interface {
	HWND() win32.HWND
	DPI() (win32.UINT, error)
	InvalidateRect(rect *win32.RECT, eraseBk bool) error
	GetClientRect() (*win32.RECT, error)
	GetWindowRect() (*win32.RECT, error)
	Value(key any) any
	SetValue(key, value any)
	SetWndProc(wndProc window.WndProc)
	TrackPopupMenu(menu *menu.Menu, spec *window.PopupMenuSpec) error
}

func DestroyWindow(handle Handle) (err error) {
	if err = win32.DestroyWindow(handle.(winBase).HWND()); err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func CreateButton(parent Handle, title string) (handle Handle, err error) {
	handle, err = button.New(parent.(winBase).HWND(), &button.Spec{
		Style: win32.WS_CHILD | win32.WS_VISIBLE,
		Text:  title,
	})
	err = errortrace.WithStack(err)
	return
}

func SetButtonOnClickListener(handle Handle, onClick func()) {
	btn := handle.(*button.Button)
	btn.OnClick = onClick
}

func SetButtonLabel(handle Handle, label string) (err error) {
	err = handle.(*button.Button).SetText(label)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func CreateLabel(parent Handle, title string) (handle Handle, err error) {
	handle, err = static.New(parent.(winBase).HWND(), &static.Spec{
		Style: win32.WS_CHILD | win32.WS_VISIBLE |
			static.SS_NOPREFIX | static.SS_CENTER | static.SS_CENTERIMAGE, // SS_CENTERIMAGE vertically centers the single line of text.
		Text: title,
	})
	err = errortrace.WithStack(err)
	return
}

func SetLabelMultiline(handle Handle, multiline bool) (err error) {
	lbl := handle.(*static.Static)
	if multiline {
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Remove: static.SS_CENTERIMAGE,
		})
	} else {
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Add: static.SS_CENTERIMAGE,
		})
	}
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

type TextAlignment int

// Sync with native.TextAlignment

const (
	Left TextAlignment = iota
	Right
	Center
)

func SetLabelTextAlignment(handle Handle, alignment TextAlignment) (err error) {
	lbl := handle.(*static.Static)
	switch alignment {
	case Left:
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Add:    static.SS_LEFT,
			Remove: static.SS_CENTER | static.SS_RIGHT,
		})
	case Right:
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Add:    static.SS_RIGHT,
			Remove: static.SS_CENTER | static.SS_LEFT,
		})
	case Center:
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Add:    static.SS_CENTER,
			Remove: static.SS_LEFT | static.SS_RIGHT,
		})
	}
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func SetLabelText(handle Handle, text string) error {
	err := handle.(*static.Static).SetText(text)
	return errortrace.WithStack(err)
}

// SetLabelBackgroundColor sets the background color of the label.
// If color is nil, the default color is used.
func SetLabelBackgroundColor(handle Handle, color *color.NRGBA) error {
	var clr win32.COLORREF
	if color != nil {
		clr = nativeColor(color)
	} else {
		clr = win32.COLORREF(win32.GetSysColor(win32.COLOR_WINDOW))
	}
	return handle.(*static.Static).SetBackgroundColor(clr)
}

func CreateTextField(parent Handle, initialValue string, password bool) (handle Handle, err error) {
	style := win32.WS_CHILD | win32.WS_VISIBLE | win32.WS_BORDER | edit.ES_LEFT
	if password {
		// Password EDIT control must be single line.
		// Remove ES_MULTILINE for safety.
		style &^= edit.ES_MULTILINE
		style |= edit.ES_PASSWORD
	}
	handle, err = edit.New(parent.(winBase).HWND(), &edit.Spec{
		Text:  initialValue,
		Style: style,
	})
	err = errortrace.WithStack(err)
	return
}

func GetTextFieldText(handle Handle) (text string, err error) {
	text, err = handle.(*edit.Edit).Text()
	err = errortrace.WithStack(err)
	return
}

func SetTextFieldText(handle Handle, text string) error {
	err := handle.(*edit.Edit).SetText(text)
	return errortrace.WithStack(err)
}

func SetWidgetDimensions(handle Handle, x, y, width, height metrics.DP) error {
	win := handle.(winBase)

	parent, err := win32.GetParent(win.HWND())
	if err != nil {
		return errortrace.WithStack(err)
	}

	// Get the rect in parent's client area before moving/resizing.
	clientBefore, err := win.GetWindowRect()
	if err != nil {
		return errortrace.WithStack(err)
	}
	if err = win32util.ScreenToClient(parent, clientBefore); err != nil {
		return errortrace.WithStack(err)
	}

	dpi, err := win32.GetDpiForWindow(win.HWND())
	if err != nil {
		return errortrace.WithStack(err)
	}
	err = win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		win32.INT(x.Px(uint(dpi))), win32.INT(y.Px(uint(dpi))),
		win32.INT(width.Px(uint(dpi))), win32.INT(height.Px(uint(dpi))),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE)
	if err != nil {
		return errortrace.WithStack(err)
	}

	// Get the rect in parent's client area after moving/resizing.
	clientAfter, err := win.GetWindowRect()
	if err != nil {
		return errortrace.WithStack(err)
	}
	if err = win32util.ScreenToClient(parent, clientAfter); err != nil {
		return errortrace.WithStack(err)
	}

	// Compute the union of the two rects.
	clientToRedraw := &win32.RECT{
		Left:   min(clientAfter.Left, clientBefore.Left),
		Top:    min(clientAfter.Top, clientBefore.Top),
		Right:  max(clientAfter.Right, clientBefore.Right),
		Bottom: max(clientAfter.Bottom, clientBefore.Bottom),
	}

	// Invalidate the areas before and after resizing.
	// This is necessary to avoid visual glitches of sibling controls
	// if the resizing or moving causes overlapping.
	return win32.InvalidateRect(parent, clientToRedraw, true)
}

func SetWidgetEnabled(handle Handle, enabled bool) (err error) {
	win32.EnableWindow(handle.(winBase).HWND(), enabled)
	return
}

func GetWidgetEnabled(handle Handle) bool {
	return win32.IsWindowEnabled(handle.(winBase).HWND())
}

func SetWidgetSize(handle Handle, width, height int) error {
	err := win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		0, 0,
		win32.INT(width), win32.INT(height),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE|win32.SWP_NOMOVE)
	return errortrace.WithStack(err)
}

func SetWindowOnSizeChangedListener(handle Handle, onSizeChanged func(width, height metrics.DP)) {
	win := handle.(*window.Window)
	win.AddMsgListener(win32.WM_SIZE, func(hwnd win32.HWND, message win32.UINT, wParam win32.WPARAM, lParam win32.LPARAM) {
		_, _, width, height, err := WindowClientRect(win)
		if err != nil {
			panic(err)
		}
		onSizeChanged(width, height)
	})
}

func SetWindowOnDestroyListener(handle Handle, onClose func()) {
	handle.(*window.Window).OnDestroy = onClose
}

func SetWindowOnCloseListener(handle Handle, onClose func() bool) {
	handle.(*window.Window).OnClose = onClose
}

func WindowClientRect(handle Handle) (x, y, width, height metrics.DP, err error) {
	win := handle.(*window.Window)
	var rect win32.RECT
	err = win32.GetClientRect(win.HWND(), &rect)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	dip, err := win32.GetDpiForWindow(win.HWND())
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	x = metrics.Px(int(rect.Left), uint(dip))
	y = metrics.Px(int(rect.Top), uint(dip))
	width = metrics.Px(int(rect.Right-rect.Left), uint(dip))
	height = metrics.Px(int(rect.Bottom-rect.Top), uint(dip))
	return
}

var getSystemMetricsXEdge = func() func() int {
	var x win32.INT = 0
	return func() int {
		if x == 0 {
			x = win32.GetSystemMetrics(win32.SystemMetricsIndex(win32.SM_CXEDGE))
		}
		return int(x)
	}
}()

var getSystemMetricsYEdge = func() func() int {
	var y win32.INT = 0
	return func() int {
		if y == 0 {
			y = win32.GetSystemMetrics(win32.SystemMetricsIndex(win32.SM_CYEDGE))
		}
		return int(y)
	}
}()

// GetTextDrawingSize returns the size required to draw the specified text
// in the given control.
// If multiline is true, the line ending characters are considered as line breaks.
// If maxWidth is greater than zero, the returned width will not exceed maxWidth.
func GetTextDrawingSize(control Handle, text string, multiline bool, maxWidth metrics.DP) (width, height metrics.DP, err error) {
	win := control.(winBase)
	hdc, err := win32.GetDC(win.HWND())
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	defer win32.ReleaseDC(win.HWND(), hdc)
	font, err := win32.SendMessageW(win.HWND(), win32.WM_GETFONT, 0, 0)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	oldFont, err := win32.SelectObject(hdc, win32.HFONT(font))
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	defer win32.SelectObject(hdc, oldFont)

	format := win32.DT_CALCRECT | win32.DT_NOPREFIX
	if !multiline {
		format |= win32.DT_SINGLELINE
	} else {
		format |= win32.DT_WORDBREAK
	}

	var buf []win32.WCHAR
	win32util.CString(text, &buf)

	dpi, err := win32.GetDpiForWindow(win.HWND())
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}

	const MAX_SIZE = 1<<(unsafe.Sizeof(win32.LONG(0))*8-1) - 1
	rect := win32.RECT{Left: 0, Top: 0,
		Right:  gg.If(maxWidth > 0, win32.LONG(maxWidth.Px(uint(dpi))), MAX_SIZE),
		Bottom: MAX_SIZE}
	_, err = win32.DrawTextExW(hdc, &buf[0], -1,
		&rect,
		format, nil)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	return metrics.Px(int(rect.Width()), uint(dpi)), metrics.Px(int(rect.Height()), uint(dpi)), nil
}

func GetButtonMinimumSize(handle Handle, label string) (width, height metrics.DP, err error) {
	btn := handle.(*button.Button)
	style, err := win32.GetWindowLongPtrW(btn.HWND(), win32.GWL_STYLE)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	width, height, err = GetTextDrawingSize(handle, label, style&win32.BS_MULTILINE != 0, 0)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	dpi, err := win32.GetDpiForWindow(btn.HWND())
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	xEdge := metrics.Px(int(getSystemMetricsXEdge()), uint(dpi))
	yEdge := metrics.Px(int(getSystemMetricsYEdge()), uint(dpi))
	return width + xEdge*2, height + yEdge*2, nil
}

func CreatePanel(parent Handle) (handle Handle, err error) {
	handle, err = panel.New(parent.(winBase).HWND(), &panel.Spec{})
	return
}

func SetPanelBackgroundColor(handle Handle, color *color.NRGBA) error {
	return handle.(*panel.Panel).SetBackgroundColor(nativeColor(color))
}

// MouseEventListener defines methods to listen to mouse events.
// Parameter win of the methods is the window whose client coordinates are used for the event positions.
type MouseEventListener interface {
	OnMousePrimaryDown(win Handle, x, y metrics.DP)
	OnMousePrimaryUp(win Handle, x, y metrics.DP)
	OnMouseSecondaryDown(win Handle, x, y metrics.DP)
	OnMouseSecondaryUp(win Handle, x, y metrics.DP)
	OnMouseMiddleDown(win Handle, x, y metrics.DP)
	OnMouseMiddleUp(win Handle, x, y metrics.DP)
	OnMousePointerMove(win Handle, x, y metrics.DP)
}

var mouseEventListeners map[*MouseEventListener]Handle

// callMouseEventListeners calls the specified method of all registered mouse event listeners.
func callMouseEventListeners(msg *win32.MSG, method func(listener MouseEventListener, parent Handle, x metrics.DP, y metrics.DP)) {
	pt := win32.POINT{
		X: win32.LONG(win32.GET_X_LPARAM(msg.LParam)),
		Y: win32.LONG(win32.GET_Y_LPARAM(msg.LParam))} // in Hwnd's client coordinates
	check.MustOK(win32.ClientToScreen(msg.Hwnd, &pt)) // to screen coordinates

	for listener, win := range mouseEventListeners {
		target := win.(winBase)
		check.MustOK(win32.ScreenToClient(target.HWND(), &pt)) // to target's client coordinates
		dpi := uint(check.Must(target.DPI()))
		x := metrics.Px(int(pt.X), dpi)
		y := metrics.Px(int(pt.Y), dpi)
		method(*listener, win, x, y)
	}
}

// App_AddMouseEventListener adds a mouse event listener to the specified window.
// It returns a function to remove the listener.
// Parameter win is the window whose client coordinates are used for the event positions.
func App_AddMouseEventListener(app App, win Handle, listener MouseEventListener) (remove func()) {
	if mouseEventListeners == nil {
		mouseEventListeners = make(map[*MouseEventListener]Handle)
		app.(*gwapp.GwApp).SetMessageDispatcher(func(msg *win32.MSG, prevProc func(msg *win32.MSG) win32.LRESULT) win32.LRESULT {
			switch msg.Message {
			case win32.WM_LBUTTONDOWN:
				callMouseEventListeners(msg, MouseEventListener.OnMousePrimaryDown)
			case win32.WM_LBUTTONUP:
				callMouseEventListeners(msg, MouseEventListener.OnMousePrimaryUp)
			case win32.WM_RBUTTONDOWN:
				callMouseEventListeners(msg, MouseEventListener.OnMouseSecondaryDown)
			case win32.WM_RBUTTONUP:
				callMouseEventListeners(msg, MouseEventListener.OnMouseSecondaryUp)
			case win32.WM_MBUTTONDOWN:
				callMouseEventListeners(msg, MouseEventListener.OnMouseMiddleDown)
			case win32.WM_MBUTTONUP:
				callMouseEventListeners(msg, MouseEventListener.OnMouseMiddleUp)
			case win32.WM_MOUSEMOVE:
				callMouseEventListeners(msg, MouseEventListener.OnMousePointerMove)
			}
			return prevProc(msg)
		})
	}
	mouseEventListeners[&listener] = win
	return func() {
		delete(mouseEventListeners, &listener)
	}
}

func ClientCoordinatesConv(from, to Handle, x, y metrics.DP) (newX, newY metrics.DP, err error) {
	fromWin := from.(winBase).HWND()
	toWin := to.(winBase).HWND()
	fromDpi, err := win32.GetDpiForWindow(fromWin)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	toDpi, err := win32.GetDpiForWindow(toWin)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	var pt win32.POINT
	pt.X = win32.LONG(x.Px(uint(fromDpi)))
	pt.Y = win32.LONG(y.Px(uint(fromDpi)))
	if err = win32.ClientToScreen(fromWin, &pt); err != nil {
		err = errortrace.WithStack(err)
		return
	}
	if err = win32.ScreenToClient(toWin, &pt); err != nil {
		err = errortrace.WithStack(err)
		return
	}
	newX = metrics.Px(int(pt.X), uint(toDpi))
	newY = metrics.Px(int(pt.Y), uint(toDpi))
	return
}

// ClientToScreenPt converts the point (x, y) in the client area of the specified window
// to screen coordinates.
func ClientToScreen(win Handle, x, y metrics.DP) (screenX, screenY metrics.DP, err error) {
	w := win.(winBase)
	dpi, err := w.DPI()
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	pt := &win32.POINT{
		X: win32.LONG(x.Px(uint(dpi))),
		Y: win32.LONG(y.Px(uint(dpi))),
	}
	if err = win32.ClientToScreen(w.HWND(), pt); err != nil {
		err = errortrace.WithStack(err)
		return
	}
	return metrics.Px(int(pt.X), uint(dpi)), metrics.Px(int(pt.Y), uint(dpi)), nil
}
