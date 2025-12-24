package native

import (
	"iter"
	"unsafe"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/gw/app/gwapp"
	"github.com/mkch/gw/button"
	"github.com/mkch/gw/edit"
	"github.com/mkch/gw/metrics"
	"github.com/mkch/gw/paint"
	"github.com/mkch/gw/paint/brush"
	"github.com/mkch/gw/paint/pen"
	"github.com/mkch/gw/static"
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
		err = errortrace.WithStack(err)
		return
	}
	win.Show(win32.SW_SHOWNORMAL)
	handle = win
	return
}

func InvalidWindow(handle Handle) error {
	err := handle.(*window.Window).InvalidateRect(nil, true)
	return errortrace.WithStack(err)
}

type winBase interface {
	HWND() win32.HWND
}

func DestroyWindow(handle Handle) error {
	err := win32.DestroyWindow(handle.(winBase).HWND())
	return errortrace.WithStack(err)
}

func CreateButton(parent Handle, title string) (handle Handle, err error) {
	handle, err = button.New(parent.(winBase).HWND(), &button.Spec{
		Style:  win32.WS_CHILD | win32.WS_VISIBLE,
		Text:   title,
		Width:  metrics.Px(100),
		Height: metrics.Px(30),
	})
	err = errortrace.WithStack(err)
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

func CreateLabel(parent Handle, title string) (handle Handle, err error) {
	handle, err = static.New(parent.(winBase).HWND(), &static.Spec{
		Style:  win32.WS_CHILD | win32.WS_VISIBLE,
		Text:   title,
		Width:  metrics.Px(100),
		Height: metrics.Px(30),
	})
	err = errortrace.WithStack(err)
	return
}

func SetLabelText(handle Handle, text string) error {
	err := handle.(*static.Static).SetText(text)
	return errortrace.WithStack(err)
}

func CreateTextField(parent Handle) (handle Handle, err error) {
	handle, err = edit.New(parent.(winBase).HWND(), &edit.Spec{
		Style:  win32.WS_CHILD | win32.WS_VISIBLE | win32.WS_BORDER | edit.ES_LEFT,
		Width:  metrics.Px(200),
		Height: metrics.Px(30),
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

func SetWidgetDimensions(handle Handle, x, y, width, height int) error {
	err := win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		win32.INT(x), win32.INT(y),
		win32.INT(width), win32.INT(height),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE)
	return errortrace.WithStack(err)
}

func SetWidgetSize(handle Handle, width, height int) error {
	err := win32.SetWindowPos(handle.(winBase).HWND(), win32.HWND(0),
		0, 0,
		win32.INT(width), win32.INT(height),
		win32.SWP_NOZORDER|win32.SWP_NOACTIVATE|win32.SWP_NOMOVE)
	return errortrace.WithStack(err)
}

func SetWindowOnSizeChangedListener(handle Handle, onSizeChanged func(width, height int)) {
	win := handle.(*window.Window)
	win.AddMsgListener(win32.WM_SIZE, func(hwnd win32.HWND, message win32.UINT, wParam win32.WPARAM, lParam win32.LPARAM) {
		width := win32.LOWORD(uintptr(lParam))
		height := win32.HIWORD(uintptr(lParam))
		onSizeChanged(int(width), int(height))
	})
}

func SetWindowOnCloseListener(handle Handle, onClose func()) {
	win := handle.(*window.Window)
	win.OnClose = onClose
}

func WindowClientRect(handle Handle) (x, y, width, height int, err error) {
	win := handle.(*window.Window)
	var rect win32.RECT
	err = win32.GetClientRect(win.HWND(), &rect)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	return int(rect.Left), int(rect.Top), int(rect.Right - rect.Left), int(rect.Bottom - rect.Top), nil
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
func GetTextDrawingSize(control Handle, text string, multiline bool) (width, height int, err error) {
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

	format := win32.DT_CALCRECT
	if !multiline {
		format |= win32.DT_SINGLELINE
	}

	var buf []win32.WCHAR
	win32util.CString(text, &buf)
	const MAX_SIZE = 1<<(unsafe.Sizeof(win32.LONG(0))*8-1) - 1
	rect := win32.RECT{Left: 0, Top: 0, Right: MAX_SIZE, Bottom: MAX_SIZE}
	_, err = win32.DrawTextExW(hdc, &buf[0], -1,
		&rect,
		format, nil)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	return int(rect.Width()), int(rect.Height()), nil
}

func GetButtonMinimumSize(handle Handle, label string) (width, height int, err error) {
	btn := handle.(*button.Button)
	style, err := win32.GetWindowLongPtrW(btn.HWND(), win32.GWL_STYLE)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	width, height, err = GetTextDrawingSize(handle, label, style&win32.BS_MULTILINE != 0)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	return int(width + int(win32.LONG(getSystemMetricsXEdge())*2)), int(height + int(win32.LONG(getSystemMetricsYEdge())*2)), nil
}

var debugRectPen = func() func() *pen.Pen {
	var p *pen.Pen
	return func() *pen.Pen {
		if p == nil {
			p = gg.Must(pen.NewCosmetic(win32.PS_DOT, win32.RGB(255, 0, 0)))
		}
		return p
	}
}()

var debugRectHollowBrush = func() func() *brush.Brush {
	var b *brush.Brush
	return func() *brush.Brush {
		if b == nil {
			b = gg.Must(brush.NewStock(win32.NULL_BRUSH))
		}
		return b
	}
}()

var debugRectHighlightBrush = func() func() *brush.Brush {
	var b *brush.Brush
	return func() *brush.Brush {
		if b == nil {
			b = gg.Must(brush.New(&win32.LOGBRUSH{
				Style: win32.BS_SOLID,
				Color: win32.RGB(255, 0, 0),
			}))
		}
		return b
	}
}()

type DebugRect struct {
	Left, Top, Right, Bottom int
	Highlight                bool
}

func EnableDrawDebugRect(winHandle Handle, rects func() iter.Seq[DebugRect]) error {
	win := winHandle.(*window.Window)
	win.SetPaintCallback(func(dc *paint.PaintDC, prev func(*paint.PaintDC)) {
		pen := debugRectPen()
		oldPen, _ := win32.SelectObject(dc.HDC(), pen.HPEN())
		defer win32.SelectObject(dc.HDC(), oldPen)

		for rect := range rects() {
			var oldBrush win32.HBRUSH
			if rect.Highlight {
				oldBrush, _ = win32.SelectObject(dc.HDC(), debugRectHighlightBrush().HBRUSH())
			} else {
				oldBrush, _ = win32.SelectObject(dc.HDC(), debugRectHollowBrush().HBRUSH())
			}
			win32.Rectangle(dc.HDC(),
				rect.Left,
				rect.Top,
				rect.Right,
				rect.Bottom)
			win32.SelectObject(dc.HDC(), oldBrush)
		}

	})
	return nil

}
