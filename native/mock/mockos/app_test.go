package mockos

import (
	"image/color"
	"testing"

	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

func Test_Run(t *testing.T) {
	const exitCode = 123
	var destroyListenerCalled = false
	code := Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		if w1 == nil {
			t.Fatalf("NewWindow returned nil handle")
		}
		SetOnDestroyListener(w1, func() { destroyListenerCalled = true })

		if err := PostQuitMessage(exitCode); err != nil {
			t.Fatalf("PostQuitMessage failed: %v", err)
		}
	})
	if code != exitCode {
		t.Fatalf("Run returned %d, want %d", code, exitCode)
	}

	// Test that windows are destroyed after Run returns.
	if len(windows) != 0 {
		t.Fatalf("windows not destroyed after Run returns")
	}
	if !destroyListenerCalled {
		t.Fatalf("destroy listener not called")
	}
}

func Test_SetMessageDispatcher(t *testing.T) {
	var dispatcherCalled = false
	var prevDispatcherCalled = false
	Run(func() {
		SetMessageDispatcher(func(msg *Msg, prev func(*Msg) any) any {
			dispatcherCalled = true
			return 123
		})
		SetMessageDispatcher(func(msg *Msg, prev func(*Msg) any) any {
			prevDispatcherCalled = true
			ret := prev(msg)
			if ret != 123 {
				t.Fatalf("previous dispatcher returned %v, want 123", ret)
			}
			return ret
		})

		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		if err := PostMessage(w1, MsgMouseLeftDown{}); err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("PostQuitMessage failed: %v", err)
		}
	})
	if !dispatcherCalled {
		t.Fatalf("custom dispatcher not called")
	}
	if !prevDispatcherCalled {
		t.Fatalf("previous dispatcher not called")
	}
}

func Test_Post(t *testing.T) {
	var fCalled = make(chan struct{})
	Run(func() {
		err := Post(func() {
			close(fCalled)
		})
		if err != nil {
			t.Fatalf("Post failed: %v", err)
		}
		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("PostQuitMessage failed: %v", err)
		}
	})
	<-fCalled // Should not block
}

func Test_ClientScreenCords(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{Width: 100, Height: 100})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		w1Rect, err := WindowRect(w1)
		if err != nil {
			t.Fatalf("WindowRect failed: %v", err)
		}
		btn, err := NewButton(w1, "button1")
		if err != nil {
			t.Fatalf("NewButton failed: %v", err)
		}
		btnRect := metrics.Rect{Left: 10, Top: 11, Right: 60, Bottom: 40}
		SetControlDimensions(btn, btnRect)

		// convert to screen coords
		screenX, screenY, err := ClientToScreen(btn, 5, 6)
		if err != nil {
			t.Fatalf("ClientToScreen failed: %v", err)
		}
		expectedScreenX := 5 + btnRect.Left + w1Rect.Left
		expectedScreenY := 6 + btnRect.Top + w1Rect.Top
		if screenX != expectedScreenX || screenY != expectedScreenY {
			t.Fatalf("ClientToScreen returned (%v, %v), want (%v, %v)",
				screenX, screenY, expectedScreenX, expectedScreenY)
		}
		// convert back
		clientX, clientY, err := ScreenToClient(btn, screenX, screenY)
		if err != nil {
			t.Fatalf("ScreenToClient failed: %v", err)
		}
		if clientX != 5 || clientY != 6 {
			t.Fatalf("ScreenToClient returned (%v, %v), want (5, 6)",
				clientX, clientY)
		}

		if err = PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_CloseWindow(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		var onCloseCalled = 0
		var onDestroyCalled = 0
		SetOnCloseListener(w1, func() bool {
			onCloseCalled++
			return false // do not close
		})
		SetOnDestroyListener(w1, func() {
			onDestroyCalled++
		})

		// Close the window
		if err := CloseWindow(w1); err != nil {
			t.Fatalf("CloseWindow failed: %v", err)
		}
		// Verify that listeners onClose is called
		if onCloseCalled != 1 {
			t.Fatalf("onCloseCalled = %d, want 1", onCloseCalled)
		}
		// Verify that window is not closed
		if _, ok := windows[w1]; !ok {
			t.Fatalf("window was closed, but onClose returned false")
		}
		// Verify that onDestroy is not called
		if onDestroyCalled != 0 {
			t.Fatalf("onDestroyCalled = %d, want 0", onDestroyCalled)
		}

		// Set onClose to return true
		SetOnCloseListener(w1, func() bool {
			onCloseCalled++
			return true // close
		})

		// Close the window
		if err := CloseWindow(w1); err != nil {
			t.Fatalf("CloseWindow failed: %v", err)
		}
		// Verify that listeners onClose is called
		if onCloseCalled != 2 {
			t.Fatalf("onCloseCalled = %d, want 1", onCloseCalled)
		}
		// Verify that window is destroyed
		if _, ok := windows[w1]; ok {
			t.Fatalf("window was closed, but onClose returned false")
		}
		// Verify that onDestroy is called
		if onDestroyCalled != 1 {
			t.Fatalf("onDestroyCalled = %d, want 1", onDestroyCalled)
		}

		if err = PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_InvalidateWindow(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		var paintCalled = 0

		expectedRect := metrics.Rect{Left: 10, Top: 11, Right: 50, Bottom: 60}
		SetWndProc(w1, func(msg *Msg, prev func(*Msg) any) any {
			switch m := msg.Message.(type) {
			case MsgPaint:
				paintCalled++
				if metrics.Rect(m) != expectedRect {
					t.Fatalf("OnPaint called with rect %v, want %v", m, expectedRect)
				}
			}
			return prev(msg)
		})
		if err := InvalidateWindow(w1, &expectedRect); err != nil {
			t.Fatalf("InvalidateWindow failed: %v", err)
		}

		if err = PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_DestroyWindow(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		var onDestroyCalled = 0
		SetOnDestroyListener(w1, func() {
			onDestroyCalled++
		})

		p, err := NewPanel(w1)
		if err != nil {
			t.Fatalf("NewPanel failed: %v", err)
		}
		SetOnDestroyListener(p, func() {
			onDestroyCalled++
		})

		if err := DestroyWindow(w1); err != nil {
			t.Fatalf("DestroyWindow failed: %v", err)
		}

		if onDestroyCalled != 2 {
			t.Fatalf("onDestroyCalled = %d, want 1", onDestroyCalled)
		}

		if windows[w1] != nil {
			t.Fatalf("window not removed from windows map")
		}
		if windows[p] != nil {
			t.Fatalf("child panel not removed from windows map")
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_SetOnSizeChangedListener(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		var sizeChangedCalled = 0
		var newSizeListened metrics.Size
		SetOnSizeChangedListener(w1, func(newSize metrics.Size) {
			sizeChangedCalled++
			newSizeListened = newSize
		})
		newSize := metrics.Size{Width: 200, Height: 150}
		rect, err := WindowRect(w1)
		if err != nil {
			t.Fatalf("WindowRect failed: %v", err)
		}
		rect.Right = rect.Left + newSize.Width
		rect.Bottom = rect.Top + newSize.Height
		if err := SetWindowRect(w1, rect); err != nil {
			t.Fatalf("SetWindowSize failed: %v", err)
		}
		if sizeChangedCalled != 1 {
			t.Fatalf("sizeChangedCalled = %d, want 1", sizeChangedCalled)
		}
		if newSizeListened != newSize {
			t.Fatalf("newSizeListened = %v, want %v", newSizeListened, newSize)
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_ClientRect(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{Width: 300, Height: 200})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		rect, err := ClientRect(w1)
		if err != nil {
			t.Fatalf("ClientRect failed: %v", err)
		}
		expectedRect := metrics.Rect{Left: 10, Top: 10, Right: 280, Bottom: 180}
		if rect != expectedRect {
			t.Fatalf("ClientRect = %v, want %v", rect, expectedRect)
		}

		err = SetWindowRect(w1, metrics.Rect{Left: 20, Top: 20, Right: 30, Bottom: 22})
		if err != nil {
			t.Fatalf("ClientRect failed: %v", err)
		}
		rect, err = ClientRect(w1)
		if err != nil {
			t.Fatalf("ClientRect failed: %v", err)
		}
		expectedRect = metrics.Rect{Left: 10, Top: 2, Right: 10, Bottom: 2}
		if rect != expectedRect {
			t.Fatalf("ClientRect = %v, want %v", rect, expectedRect)
		}

		err = SetWindowRect(w1, metrics.Rect{Left: 2, Top: 2, Right: 2, Bottom: 2})
		if err != nil {
			t.Fatalf("ClientRect failed: %v", err)
		}
		rect, err = ClientRect(w1)
		if err != nil {
			t.Fatalf("ClientRect failed: %v", err)
		}
		expectedRect = metrics.Rect{Left: 0, Top: 0, Right: 0, Bottom: 0}
		if rect != expectedRect {
			t.Fatalf("ClientRect = %v, want %v", rect, expectedRect)
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_WindowMenu(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		mu := NewMenu()
		if err = SetMenu(w1, mu); err != nil {
			t.Fatalf("SetMenu failed: %v", err)
		}
		if menuHandle, err := WindowMenu(w1); err != nil {
			t.Fatalf("WindowMenu failed: %v", err)
		} else if menuHandle != mu {
			t.Fatalf("WindowMenu = %v, want %v", menuHandle, mu)
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
	if len(menus) != 0 {
		t.Fatalf("menus not destroyed after Run returns")
	}
}

func Test_ButtonClick(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		btn, err := NewButton(w1, "button1")
		if err != nil {
			t.Fatalf("NewButton failed: %v", err)
		}

		var clicked bool
		SetButtonOnClickListener(btn, func() {
			clicked = true
		})

		if err := Debug_SimulateButtonClick(btn); err != nil {
			t.Fatalf("Debug_SimulateButtonClick failed: %v", err)
		}

		if !clicked {
			t.Fatalf("button click listener not called")
		}
		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
}

func Test_SetText(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		text, err := Text(w1)
		if err != nil {
			t.Fatalf("Text failed: %v", err)
		}
		if text != "window1" {
			t.Fatalf("Text = %q, want %q", text, "window1")
		}

		if err := SetText(w1, "Hello, World!"); err != nil {
			t.Fatalf("SetText failed: %v", err)
		}
		text, err = Text(w1)
		if err != nil {
			t.Fatalf("Text failed: %v", err)
		}
		if text != "Hello, World!" {
			t.Fatalf("Text = %q, want %q", text, "Hello, World!")
		}

		btn, err := NewButton(w1, "button1")
		if err != nil {
			t.Fatalf("NewButton failed: %v", err)
		}
		text, err = Text(btn)
		if err != nil {
			t.Fatalf("Text failed: %v", err)
		}
		if text != "button1" {
			t.Fatalf("Text = %q, want %q", text, "button1")
		}
		if err := SetText(btn, "Click Me"); err != nil {
			t.Fatalf("SetText failed: %v", err)
		}
		text, err = Text(btn)
		if err != nil {
			t.Fatalf("Text failed: %v", err)
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})

	if len(windows) != 0 {
		t.Fatalf("windows not destroyed after Run returns")
	}
}

func TestLabel(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		lbl, err := NewLabel(w1, "label1")
		if err != nil {
			t.Fatalf("NewLabel failed: %v", err)
		}
		text, err := Text(lbl)
		if err != nil {
			t.Fatalf("Text failed: %v", err)
		}
		if text != "label1" {
			t.Fatalf("Text = %q, want %q", text, "label1")
		}
		if err := SetText(lbl, "New Label Text"); err != nil {
			t.Fatalf("SetText failed: %v", err)
		}
		text, err = Text(lbl)
		if err != nil {
			t.Fatalf("Text failed: %v", err)
		}
		if text != "New Label Text" {
			t.Fatalf("Text = %q, want %q", text, "New Label Text")
		}

		err = SetLabelMultiline(lbl, true)
		if err != nil {
			t.Fatalf("SetLabelMultiline failed: %v", err)
		}
		if windows[lbl].(*label).multiline != true {
			t.Fatalf("Label multiline not set to true")
		}

		err = SetLabelMultiline(lbl, false)
		if err != nil {
			t.Fatalf("SetLabelMultiline failed: %v", err)
		}
		if windows[lbl].(*label).multiline != false {
			t.Fatalf("Label multiline not set to false")
		}

		err = SetLabelTextAlignment(lbl, native.Center)
		if err != nil {
			t.Fatalf("SetLabelTextAlignment failed: %v", err)
		}
		if windows[lbl].(*label).textAlignment != native.Center {
			t.Fatalf("Label text alignment not set to Center")
		}

		err = SetLabelBackgroundColor(lbl, &color.NRGBA{R: 255})
		if err != nil {
			t.Fatalf("SetLabelBackgroundColor failed: %v", err)
		}
		if windows[lbl].(*label).backgroundColor != (color.NRGBA{R: 255}) {
			t.Fatalf("Label background color not set to {R:255}")
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
	if len(windows) != 0 {
		t.Fatalf("windows not destroyed after Run returns")
	}
}

func Test_Enabled(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		btn, err := NewButton(w1, "button1")
		if err != nil {
			t.Fatalf("NewButton failed: %v", err)
		}
		var clicked int
		SetButtonOnClickListener(btn, func() {
			clicked++
		})

		enabled, err := ControlEnabled(btn)
		if err != nil {
			t.Fatalf("ControlEnabled failed: %v", err)
		}
		if !enabled {
			t.Fatalf("ControlEnabled = %v, want true", enabled)
		}

		err = Debug_SimulateButtonClick(btn)
		if err != nil {
			t.Fatalf("Debug_SimulateButtonClick failed: %v", err)
		}
		if clicked != 1 {
			t.Fatalf("clicked = %d, want 1", clicked)
		}

		if err := SetControlEnabled(btn, false); err != nil {
			t.Fatalf("SetControlEnabled failed: %v", err)
		}
		enabled, err = ControlEnabled(btn)
		if err != nil {
			t.Fatalf("ControlEnabled failed: %v", err)
		}
		if enabled {
			t.Fatalf("ControlEnabled = %v, want false", enabled)
		}
		err = Debug_SimulateButtonClick(btn)
		if err != nil {
			t.Fatalf("Debug_SimulateButtonClick failed: %v", err)
		}
		if clicked != 1 {
			t.Fatalf("clicked = %d, want 1", clicked)
		}

		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
	if len(windows) != 0 {
		t.Fatalf("windows not destroyed after Run returns")
	}
}

func Test_Panel(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		pnl, err := NewPanel(w1)
		if err != nil {
			t.Fatalf("NewPanel failed: %v", err)
		}
		err = SetPanelBackgroundColor(pnl, &color.NRGBA{B: 255})
		if err != nil {
			t.Fatalf("SetPanelBackgroundColor failed: %v", err)
		}
		if windows[pnl].(*panel).backgroundColor != (color.NRGBA{B: 255}) {
			t.Fatalf("Panel background color not set to {B:255}")
		}
		if err := PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
	if len(windows) != 0 {
		t.Fatalf("windows not destroyed after Run returns")
	}
}

func Test_TextDrawingSize(t *testing.T) {
	Run(func() {
		w1, err := NewWindow("window1", metrics.Size{})
		if err != nil {
			t.Fatalf("NewWindow failed: %v", err)
		}
		text := "Hello,\nWorld!"
		size1, err := TextDrawingSize(w1, text, true, 10)
		if err != nil {
			t.Fatalf("TextDrawingSize failed: %v", err)
		}
		if size1.Width == 0 || size1.Height == 0 || size1.Width > 10 {
			t.Fatalf("TextDrawingSize returned invalid size: %v", size1)
		}
		size2, err := TextDrawingSize(w1, text, false, 10)
		if err != nil {
			t.Fatalf("TextDrawingSize failed: %v", err)
		}
		if size2.Width == 0 || size2.Height == 0 || size2.Width > 10 {
			t.Fatalf("TextDrawingSize returned invalid size: %v", size2)
		}
		if size2.Width < size1.Width || size2.Height > size1.Height {
			t.Fatalf("TextDrawingSize with multiline returned unexpected size: %v vs %v", size1, size2)
		}
		if err = PostQuitMessage(0); err != nil {
			t.Fatalf("Quit failed: %v", err)
		}
	})
	if len(windows) != 0 {
		t.Fatalf("windows not destroyed after Run returns")
	}
}
