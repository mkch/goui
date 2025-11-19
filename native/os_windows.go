package native

import (
	"unsafe"

	"github.com/mkch/gw/app/gwapp"
	"github.com/mkch/gw/button"
	"github.com/mkch/gw/metrics"
	"github.com/mkch/gw/win32"
	"github.com/mkch/gw/win32/win32util"
	"github.com/mkch/gw/window"
)

type App = *gwapp.GwApp

func NewApp() App {
	return gwapp.New()
}

// Handle represents a platform-specific GUI object.
type Handle any

// CreateWindow creates a native window with the specified configuration.
func CreateWindow(title string, width, height int) (handle Handle, err error) {
	win, err := window.New(&window.Spec{
		Text:   title,
		Style:  win32.WS_OVERLAPPEDWINDOW | win32.WS_VISIBLE,
		X:      metrics.Px(win32.CW_USEDEFAULT),
		Width:  metrics.Px(win32.INT(width)),
		Height: metrics.Px(win32.INT(height)),
	})
	if err != nil {
		return
	}
	win.Show(win32.SW_SHOWNORMAL)
	handle = win
	return
}

type winBase interface {
	HWND() win32.HWND
}

func DestroyWindow(handle Handle) error {
	return win32.DestroyWindow(handle.(winBase).HWND())
}

func CreateButton(parent Handle, title string) (handle Handle, err error) {
	btn, err := button.New(parent.(winBase).HWND(), &button.Spec{
		Style:  win32.WS_CHILD | win32.WS_VISIBLE,
		Text:   title,
		Width:  metrics.Px(100),
		Height: metrics.Px(30),
	})
	if err != nil {
		return
	}
	handle = btn
	return
}

func SetButtonOnClickListener(handle Handle, onClick func()) {
	btn := handle.(*button.Button)
	btn.OnClick = onClick
}

func SetButtonLabel(handle Handle, label string) {
	btn := handle.(*button.Button)
	btn.SetText(label)
}

func SetWidgetDimensions(handle Handle, x, y, width, height int) error {
	return win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		win32.INT(x), win32.INT(y),
		win32.INT(width), win32.INT(height),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE)
}

func SetWidgetSize(handle Handle, width, height int) error {
	return win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		0, 0,
		win32.INT(width), win32.INT(height),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE|win32.SWP_NOMOVE)
}

func SetWidgetPosition(handle Handle, x, y int) error {
	return win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		win32.INT(x), win32.INT(y),
		0, 0,
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE|win32.SWP_NOSIZE)
}

func SetWindowOnSizeChangedListener(handle Handle, onSizeChanged func(width, height int)) {
	win := handle.(*window.Window)
	win.AddMsgListener(win32.WM_SIZE, func(hwnd win32.HWND, message win32.UINT, wParam win32.WPARAM, lParam win32.LPARAM) {
		width := win32.LOWORD(uintptr(lParam))
		height := win32.HIWORD(uintptr(lParam))
		onSizeChanged(int(width), int(height))
	})
}

func WindowClientRect(handle Handle) (x, y, width, height int, err error) {
	win := handle.(*window.Window)
	var rect win32.RECT
	err = win32.GetClientRect(win.HWND(), &rect)
	if err != nil {
		return
	}
	return int(rect.Left), int(rect.Top), int(rect.Right - rect.Left), int(rect.Bottom - rect.Top), nil
}

var GetSystemMetricsXEdge = func() func() int {
	var x win32.INT = 0
	return func() int {
		if x == 0 {
			x = win32.GetSystemMetrics(win32.SystemMetricsIndex(win32.SM_CXEDGE))
		}
		return int(x)
	}
}()

var GetSystemMetricsYEdge = func() func() int {
	var y win32.INT = 0
	return func() int {
		if y == 0 {
			y = win32.GetSystemMetrics(win32.SystemMetricsIndex(win32.SM_CYEDGE))
		}
		return int(y)
	}
}()

func GetButtonMinimumSize(handle Handle, label string) (width, height int, err error) {
	btn := handle.(*button.Button)
	hdc, err := win32.GetDC(btn.HWND())
	if err != nil {
		return
	}
	font, err := win32.SendMessageW(btn.HWND(), win32.WM_GETFONT, 0, 0)
	if err != nil {
		return
	}
	oldFont, err := win32.SelectObject(hdc, win32.HFONT(font))
	if err != nil {
		return
	}
	defer func() { win32.SelectObject(hdc, oldFont); win32.ReleaseDC(btn.HWND(), hdc) }()

	var buf []win32.WCHAR
	win32util.CString(label, &buf)
	style, err := win32.GetWindowLongPtrW(btn.HWND(), win32.GWL_STYLE)
	if err != nil {
		return
	}
	format := win32.DT_CALCRECT
	if style&win32.BS_MULTILINE == 0 {
		format |= win32.DT_SINGLELINE
	}

	const MAX_SIZE = 1<<(unsafe.Sizeof(win32.LONG(0))*8-1) - 1
	rect := win32.RECT{Left: 0, Top: 0, Right: MAX_SIZE, Bottom: MAX_SIZE}
	_, err = win32.DrawTextExW(hdc, &buf[0], -1,
		&rect,
		format, nil)
	if err != nil {
		return
	}
	return int(rect.Width() + win32.LONG(GetSystemMetricsXEdge())*2), int(rect.Height() + win32.LONG(GetSystemMetricsYEdge())*2), nil
}
