package windows

import (
	"image/color"
	"unsafe"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
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

type OS struct {
	app *gwapp.GwApp
}

func NewOS() *OS {
	return &OS{}
}

func (os *OS) App_Run(f func()) (ret int) {
	defer func() { os.app = nil }()
	os.app = gwapp.New()
	f()
	return os.app.Run()
}

func (os *OS) App_Post(f func()) error {
	return os.app.Post(f)
}

func (os *OS) App_Quit(exitCode int) error {
	os.app.Quit(exitCode)
	return nil
}

func (OS) NewWindow(title string, size metrics.Size) (handle native.Handle, err error) {
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
		win32.INT(size.Width.Px(uint(dpi))), win32.INT(size.Height.Px(uint(dpi))),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE|win32.SWP_NOMOVE)
	win.Show(win32.SW_SHOWNORMAL)
	handle = win
	return
}

func (OS) Window_Close(handle native.Handle) error {
	_, err := win32.SendMessageW(handle.(*window.Window).HWND(), win32.WM_CLOSE, 0, 0)
	return err
}

func (OS) Window_Invalidate(handle native.Handle, rect *metrics.Rect) (err error) {
	var win32Rect *win32.RECT
	if rect != nil {
		win32Rect = &win32.RECT{
			Left:   win32.LONG(rect.Left),
			Top:    win32.LONG(rect.Top),
			Right:  win32.LONG(rect.Right),
			Bottom: win32.LONG(rect.Bottom),
		}
	}
	err = handle.(winBase).InvalidateRect(win32Rect, true)
	if err != nil {
		return errortrace.WithStack(err)
	}
	return
}

func (OS) Window_Destroy(handle native.Handle) (err error) {
	if err = win32.DestroyWindow(handle.(winBase).HWND()); err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (os OS) Control_Destroy(handle native.Handle) (err error) {
	return os.Window_Destroy(handle)
}

func (OS) NewButton(parent native.Handle, title string) (handle native.Handle, err error) {
	handle, err = button.New(parent.(winBase).HWND(), &button.Spec{
		Style: win32.WS_CHILD | win32.WS_VISIBLE,
		Text:  title,
	})
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) Button_SetOnClickListener(handle native.Handle, onClick func()) error {
	btn := handle.(*button.Button)
	btn.OnClick = onClick
	return nil
}

func (OS) Button_SetLabel(handle native.Handle, label string) (err error) {
	err = handle.(*button.Button).SetText(label)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) NewLabel(parent native.Handle, title string) (handle native.Handle, err error) {
	handle, err = static.New(parent.(winBase).HWND(), &static.Spec{
		Style: win32.WS_CHILD | win32.WS_VISIBLE |
			static.SS_NOPREFIX | static.SS_CENTER | static.SS_CENTERIMAGE, // SS_CENTERIMAGE vertically centers the single line of text.
		Text: title,
	})
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) Label_SetMultiline(handle native.Handle, multiline bool) (err error) {
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

func (OS) Label_SetTextAlignment(handle native.Handle, alignment native.TextAlignment) (err error) {
	lbl := handle.(*static.Static)
	switch alignment {
	case native.Left:
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Add:    static.SS_LEFT,
			Remove: static.SS_CENTER | static.SS_RIGHT,
		})
	case native.Right:
		err = win32util.ModifyWindowStyle(lbl.HWND(), win32util.ModifyStyleSpec{
			Add:    static.SS_RIGHT,
			Remove: static.SS_CENTER | static.SS_LEFT,
		})
	case native.Center:
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

func (OS) Label_SetText(handle native.Handle, text string) error {
	err := handle.(*static.Static).SetText(text)
	if err != nil {
		return errortrace.WithStack(err)
	}
	return nil
}

// nativeColor converts a color.NRGBA to a native OS color representation.
func nativeColor(c *color.NRGBA) win32.COLORREF {
	return win32.RGB(c.R, c.G, c.B)
}

// SetLabelBackgroundColor sets the background color of the label.
// If color is nil, the default color is used.
func (OS) Label_SetBackgroundColor(handle native.Handle, color *color.NRGBA) error {
	var clr win32.COLORREF
	if color != nil {
		clr = nativeColor(color)
	} else {
		clr = win32.COLORREF(win32.GetSysColor(win32.COLOR_WINDOW))
	}
	return handle.(*static.Static).SetBackgroundColor(clr)
}

func (OS) NewTextField(parent native.Handle, initialValue string, password bool) (handle native.Handle, err error) {
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
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) TextField_Text(handle native.Handle) (text string, err error) {
	text, err = handle.(*edit.Edit).Text()
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) TextField_SetText(handle native.Handle, text string) error {
	err := handle.(*edit.Edit).SetText(text)
	if err != nil {
		return errortrace.WithStack(err)
	}
	return nil
}

func (OS) Control_SetDimensions(handle native.Handle, rect metrics.Rect) error {
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

	// Get the rect in parent's client area after moving/resizing.
	clientAfter := win32.RECT{
		Left:   win32.LONG(rect.Left.Px(uint(dpi))),
		Top:    win32.LONG(rect.Top.Px(uint(dpi))),
		Right:  win32.LONG(rect.Right.Px(uint(dpi))),
		Bottom: win32.LONG(rect.Bottom.Px(uint(dpi))),
	}

	if clientAfter == *clientBefore {
		return nil
	}

	err = win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		win32.INT(clientAfter.Left), win32.INT(clientAfter.Top),
		win32.INT(clientAfter.Width()), win32.INT(clientAfter.Height()),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE)
	if err != nil {
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

func (OS) Control_SetEnabled(handle native.Handle, enabled bool) (err error) {
	win32.EnableWindow(handle.(winBase).HWND(), enabled)
	return
}

func (OS) Control_Enabled(handle native.Handle) (bool, error) {
	return win32.IsWindowEnabled(handle.(winBase).HWND()), nil
}

func (os OS) Window_SetOnSizeChangedListener(handle native.Handle, onSizeChanged func(size metrics.Size)) error {
	win := handle.(*window.Window)
	win.AddMsgListener(win32.WM_SIZE, func(hwnd win32.HWND, message win32.UINT, wParam win32.WPARAM, lParam win32.LPARAM) {
		rect, err := os.Window_ClientRect(handle)
		if err != nil {
			panic(err)
		}
		onSizeChanged(rect.Size())
	})
	return nil
}

func (OS) Window_SetOnDestroyListener(handle native.Handle, listener func()) error {
	handle.(*window.Window).OnDestroy = listener
	return nil
}

func (OS) Control_SetOnDestroyListener(handle native.Handle, listener func()) error {
	handle.(*window.Window).OnDestroy = listener
	return nil
}

func (OS) Window_SetOnCloseListener(handle native.Handle, listener func() bool) error {
	handle.(*window.Window).OnClose = listener
	return nil
}

func (OS) Window_ClientRect(handle native.Handle) (ret metrics.Rect, err error) {
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
	ret = metrics.Rect{
		Left:   metrics.Px(int(rect.Left), uint(dip)),
		Top:    metrics.Px(int(rect.Top), uint(dip)),
		Right:  metrics.Px(int(rect.Right), uint(dip)),
		Bottom: metrics.Px(int(rect.Bottom), uint(dip)),
	}
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

func (OS) Control_TextDrawingSize(control native.Handle, text string, multiline bool, maxWidth metrics.DP) (size metrics.Size, err error) {
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
	return metrics.Size{
			Width:  metrics.Px(int(rect.Width()), uint(dpi)),
			Height: metrics.Px(int(rect.Height()), uint(dpi)),
		},
		nil
}

func (os OS) Button_MinimumSize(handle native.Handle, label string) (size metrics.Size, err error) {
	btn := handle.(*button.Button)
	style, err := win32.GetWindowLongPtrW(btn.HWND(), win32.GWL_STYLE)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	size, err = os.Control_TextDrawingSize(handle, label, style&win32.BS_MULTILINE != 0, 0)
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
	size.Width = size.Width + xEdge*2
	size.Height = size.Height + yEdge*2
	return
}

func (OS) NewPanel(parent native.Handle) (handle native.Handle, err error) {
	handle, err = panel.New(parent.(winBase).HWND(), &panel.Spec{})
	return
}

func (OS) Panel_SetBackgroundColor(handle native.Handle, color *color.NRGBA) error {
	return handle.(*panel.Panel).SetBackgroundColor(nativeColor(color))
}

type eventListenerRecord struct {
	listenerKey    window.MessageRetListenerKey
	eventListeners map[*native.MouseEventListener]native.Handle
}

var evtListeners eventListenerRecord

// callMouseEventListeners calls the specified method of all registered mouse event listeners.
// screenPt is the mouse position in screen coordinates.
func callMouseEventListeners(srcHwnd win32.HWND, screenPt *win32.POINT, method func(listener native.MouseEventListener, parent native.Handle, x metrics.DP, y metrics.DP)) {
	for listener, win := range evtListeners.eventListeners {
		target := win.(winBase)
		// ignore error of GetAncestor because mouse events are posted to the message queue
		// and the window may have been destroyed when the event is processed.
		if root, _ := win32.GetAncestor(srcHwnd, win32.GA_ROOT); root != target.HWND() {
			// Only capture messages in target, including target itself an all its descendants.
			continue
		}
		pt := *screenPt
		chkerr.MustOK(win32.ScreenToClient(target.HWND(), &pt)) // to target's client coordinates
		dpi := uint(chkerr.Must(target.DPI()))
		x := metrics.Px(int(pt.X), dpi)
		y := metrics.Px(int(pt.Y), dpi)
		method(*listener, win, x, y)
	}
}

func makeScreenPoint(hwnd win32.HWND, lParam win32.LPARAM) *win32.POINT {
	var pt = win32.POINT{X: win32.LONG(win32.GET_X_LPARAM(lParam)), Y: win32.LONG(win32.GET_Y_LPARAM(lParam))}
	// ignore error because mouse events are posted to the message queue
	// and the window may have been destroyed when the event is processed.
	win32.ClientToScreen(hwnd, &pt) // to screen coordinates
	return &pt
}

func parseScreenPoint(lParam win32.LPARAM) *win32.POINT {
	return &win32.POINT{X: win32.LONG(win32.GET_X_LPARAM(lParam)), Y: win32.LONG(win32.GET_Y_LPARAM(lParam))}
}

func (os *OS) App_AddMouseEventListener(win native.Handle, listener native.MouseEventListener) (remove func()) {
	if evtListeners.eventListeners == nil {
		evtListeners.eventListeners = make(map[*native.MouseEventListener]native.Handle)
	}
	if evtListeners.listenerKey == (window.MessageRetListenerKey{}) {
		evtListeners.listenerKey = window.AddMessageRetListener(func(hwnd win32.HWND, message win32.UINT, wParam win32.WPARAM, lParam win32.LPARAM, result win32.LRESULT) {
			switch message {
			case win32.WM_LBUTTONDOWN:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMousePrimaryDown)
			case win32.WM_LBUTTONUP:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMousePrimaryUp)
			case win32.WM_RBUTTONDOWN:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMouseSecondaryDown)
			case win32.WM_RBUTTONUP:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMouseSecondaryUp)
			case win32.WM_MBUTTONDOWN:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMouseMiddleDown)
			case win32.WM_MBUTTONUP:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMouseMiddleUp)
			case win32.WM_MOUSEMOVE:
				callMouseEventListeners(hwnd, makeScreenPoint(hwnd, lParam), native.MouseEventListener.OnMousePointerMove)

			case win32.WM_NCLBUTTONDOWN:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMousePrimaryDown)
			case win32.WM_NCLBUTTONUP:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMousePrimaryUp)
			case win32.WM_NCRBUTTONDOWN:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMouseSecondaryDown)
			case win32.WM_NCRBUTTONUP:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMouseSecondaryUp)
			case win32.WM_NCMBUTTONDOWN:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMouseMiddleDown)
			case win32.WM_NCMBUTTONUP:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMouseMiddleUp)
			case win32.WM_NCMOUSEMOVE:
				callMouseEventListeners(hwnd, parseScreenPoint(lParam), native.MouseEventListener.OnMousePointerMove)
			}
		})
	}
	evtListeners.eventListeners[&listener] = win
	return func() {
		delete(evtListeners.eventListeners, &listener)
		if len(evtListeners.eventListeners) == 0 {
			window.RemoveMessageRetListener(evtListeners.listenerKey)
			evtListeners.listenerKey = window.MessageRetListenerKey{}
		}
	}
}

func (OS) Util_ClientCoordinatesConv(from, to native.Handle, x, y metrics.DP) (newX, newY metrics.DP, err error) {
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

func (OS) Util_ClientToScreen(win native.Handle, x, y metrics.DP) (screenX, screenY metrics.DP, err error) {
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

func (OS) Window_Menu(win native.Handle) (native.Handle, error) {
	return win32.GetMenu(win.(winBase).HWND())
}

func (OS) Window_SetMenu(win native.Handle, m native.Handle) (err error) {
	//nativeMenu.SetPopup(false) // Window menu should not be popup
	err = win.(*window.Window).SetMenu(m.(*menu.Menu))
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) Window_RefreshMenu(win native.Handle) (err error) {
	err = win32.DrawMenuBar(win.(*window.Window).HWND())
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) Window_Enabled(win native.Handle) (bool, error) {
	return win32.IsWindowEnabled(win.(winBase).HWND()), nil
}

func (OS) Window_SetEnabled(win native.Handle, enabled bool) error {
	win32.EnableWindow(win.(winBase).HWND(), enabled)
	return nil
}

func (OS) NewMenu(popup bool) (native.Handle, error) {
	return menu.New(popup), nil
}

func (OS) Menu_Destroy(m native.Handle) error {
	return m.(*menu.Menu).Destroy()
}

func (OS) NewMenuItem(parent native.Handle, title string, separator bool) (handle native.Handle, err error) {
	handle, err = parent.(*menu.Menu).InsertItem(-1, &menu.ItemSpec{
		Title:     title,
		Separator: separator,
	})
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) MenuItem_Destroy(item native.Handle) (err error) {
	nativeItem := item.(*menu.Item)
	err = nativeItem.Menu().DeleteItem(nativeItem)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) MenuItem_SetTitle(item native.Handle, title string) (err error) {
	err = item.(*menu.Item).SetTitle(title)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) MenuItem_SetDisabled(item native.Handle, disabled bool) (err error) {
	err = item.(*menu.Item).SetDisabled(disabled)
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) MenuItem_SetSubmenu(item native.Handle, submenu native.Handle) (err error) {
	if submenu == nil {
		err = item.(*menu.Item).SetSubmenu(nil)
	} else {
		err = item.(*menu.Item).SetSubmenu(submenu.(*menu.Menu))
	}
	if err != nil {
		err = errortrace.WithStack(err)
	}
	return
}

func (OS) MenuItem_SetOnClickListener(item native.Handle, listener func()) {
	item.(*menu.Item).OnClick = listener
}

func (OS) Window_TrackPopupMenu(win native.Handle, menuToTrack native.Handle, spec *native.TrackPopupSpec) (err error) {
	nativeWin := win.(winBase)
	dpi, err := nativeWin.DPI()
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	var nativeSpec *window.PopupMenuSpec
	if spec != nil {
		nativeSpec = &window.PopupMenuSpec{
			X: win32.LONG(spec.X.Px(uint(dpi))),
			Y: win32.LONG(spec.Y.Px(uint(dpi))),
		}
	}

	return nativeWin.TrackPopupMenu(menuToTrack.(*menu.Menu), nativeSpec)
}

func (OS) MessageBox(parent native.Handle, title, message string, icon native.MessageBoxIcon, button native.MessageBoxButton) (ret native.MessageBoxReturn, err error) {
	var nativeParent win32.HWND
	if parent != nil {
		nativeParent = parent.(winBase).HWND()
	}
	var nativeType win32.MESSAGE_BOX_TYPE
	switch button {
	case native.MessageBoxButtonYesNo:
		nativeType |= win32.MB_YESNO
	}
	switch icon {
	case native.MessageBoxIconInfo:
		nativeType |= win32.MB_ICONINFORMATION
	case native.MessageBoxIconWarning:
		nativeType |= win32.MB_ICONWARNING
	case native.MessageBoxIconQuestion:
		nativeType |= win32.MB_ICONQUESTION
	case native.MessageBoxIconError:
		nativeType |= win32.MB_ICONERROR
	}
	id, err := win32util.MessageBox(nativeParent, message, title, nativeType)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	switch id {
	case win32.IDOK:
		ret = native.MessageBoxReturnOK
	case win32.IDYES:
		ret = native.MessageBoxReturnYes
	case win32.IDNO:
		ret = native.MessageBoxReturnNo
	case win32.IDCANCEL:
		ret = native.MessageBoxReturnCancel
	}
	return
}
