package native

import (
	"iter"

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/gw/paint"
	"github.com/mkch/gw/paint/brush"
	"github.com/mkch/gw/paint/pen"
	"github.com/mkch/gw/panel"
	"github.com/mkch/gw/win32"
	"github.com/mkch/gw/window"
)

var debugRectPen = func() func() *pen.Pen {
	var p *pen.Pen
	return func() *pen.Pen {
		if p == nil {
			p = check.Must(pen.New(pen.NewLogPen(&win32.LOGPEN{
				Style: win32.PS_DOT,
				Width: 2,
				Color: win32.RGB(255, 0, 0),
			}, 96), 96))
			//p = check.Must(pen.NewCosmetic(win32.PS_DOT, win32.RGB(255, 0, 0)))
		}
		return p
	}
}()

var debugRectHollowBrush = func() func() *brush.Brush {
	var b *brush.Brush
	return func() *brush.Brush {
		if b == nil {
			b = check.Must(brush.NewStock(win32.NULL_BRUSH))
		}
		return b
	}
}()

var debugRectHighlightBrush = func() func() *brush.Brush {
	var b *brush.Brush
	return func() *brush.Brush {
		if b == nil {
			b = check.Must(brush.New(&win32.LOGBRUSH{
				Style: win32.BS_SOLID,
				Color: win32.RGB(255, 0, 0),
			}))
		}
		return b
	}
}()

type DebugRect struct {
	Left, Top, Right, Bottom metrics.DP
	Highlight                bool
}

func EnableDrawDebugRect(winHandle Handle, rects func() iter.Seq[DebugRect]) (layer Handle, err error) {
	win := winHandle.(*window.Window)
	client, err := win.GetClientRect()
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}
	layeredPanel, err := panel.New(win.HWND(), &panel.Spec{
		ExStyle: win32.WS_EX_LAYERED | win32.WS_EX_TRANSPARENT,
	})
	if layeredPanel == nil {
		// Layered child windows are only supported on Windows 8 or greater.
		// CreateWindowEx fails silently and returns NULL handle with a 0 last error code if
		// the application is not marked as Windows 8 or greater aware in its manifest.
		panic(
			`debug.LayoutOutline is not supported. Missing or wrong manifest? Please make sure your application manifest includes the following snippet:

<!--
https://learn.microsoft.com/en-us/windows/win32/winmsg/using-windows#using-layered-windows
In order to use layered child windows, the application has to declare itself Windows 8-aware in the manifest.
For windows 10/11, one can include this compatibility snippet in its app.manifest to make it Windows 10-aware :
-->
<compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
	<application>
		<!-- Windows 10 GUID -->
		<supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}" />
	</application>
</compatibility>`)
	}
	layer = layeredPanel

	err = win32.SetWindowPos(layeredPanel.HWND(), 0,
		0, 0, win32.INT(client.Width()), win32.INT(client.Height()),
		win32.SWP_NOACTIVATE|win32.SWP_NOZORDER)

	if err != nil {
		err = errortrace.WithStack(err)
		return
	}

	err = win32.SetLayeredWindowAttributes(layeredPanel.HWND(), layeredPanel.BackgroundColor(), 128, win32.LWA_COLORKEY|win32.LWA_ALPHA)
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}

	layeredPanel.AddPaintCallback(func(paintData *paint.PaintData, prev func(*paint.PaintData)) {
		prev(&paint.PaintData{
			DC:    paintData.DC,
			Rect:  paintData.Rect,
			Erase: paintData.Erase,
		})

		dpi := check.Must(win32.GetDpiForWindow(layeredPanel.HWND()))

		pen := debugRectPen()
		defer check.Must(paint.SelectObject(paintData.DC, pen.HPEN())).Restore()

		for rect := range rects() {
			var restore func()
			if rect.Highlight {
				restore = check.Must(paint.SelectObject(paintData.DC, debugRectHighlightBrush().HBRUSH())).Restore
			} else {
				restore = check.Must(paint.SelectObject(paintData.DC, debugRectHollowBrush().HBRUSH())).Restore
			}
			left := rect.Left.Px(uint(dpi))
			right := rect.Right.Px(uint(dpi))
			top := rect.Top.Px(uint(dpi))
			bottom := rect.Bottom.Px(uint(dpi))
			win32.Rectangle(paintData.DC,
				left, top, right, bottom)
			restore()
		}
	})

	err = layeredPanel.AddDoubleBufferingPaintCallback()
	if err != nil {
		err = errortrace.WithStack(err)
		return
	}

	win.AddMsgListener(win32.WM_SIZE, func(hwnd win32.HWND, message win32.UINT, wParam win32.WPARAM, lParam win32.LPARAM) {
		width := int(win32.LOWORD(uintptr(lParam)))
		height := int(win32.HIWORD(uintptr(lParam)))
		win32.SetWindowPos(layeredPanel.HWND(), 0,
			0, 0, win32.INT(width), win32.INT(height),
			win32.SWP_NOACTIVATE|win32.SWP_NOZORDER)
	})

	return

}
