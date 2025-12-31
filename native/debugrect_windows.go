package native

import (
	"iter"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
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
			p = gg.Must(pen.New(pen.NewLogPen(&win32.LOGPEN{
				Style: win32.PS_DOT,
				Width: 2,
				Color: win32.RGB(255, 0, 0),
			}, 96), 96))
			//p = gg.Must(pen.NewCosmetic(win32.PS_DOT, win32.RGB(255, 0, 0)))
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
	if err != nil {
		err = errortrace.WithStack(err)
		return
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

		pen := debugRectPen()
		defer gg.Must(paint.SelectObject(paintData.DC, pen.HPEN())).Restore()

		for rect := range rects() {
			var restore func()
			if rect.Highlight {
				restore = gg.Must(paint.SelectObject(paintData.DC, debugRectHighlightBrush().HBRUSH())).Restore
			} else {
				restore = gg.Must(paint.SelectObject(paintData.DC, debugRectHollowBrush().HBRUSH())).Restore
			}
			win32.Rectangle(paintData.DC,
				rect.Left,
				rect.Top,
				rect.Right,
				rect.Bottom)
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
