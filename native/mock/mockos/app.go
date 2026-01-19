package mockos

import (
	"errors"
	"image/color"
	"iter"
	"maps"
	"slices"

	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

const msgQueueSize = 1024

// windows is a map of all created [window]s.
var windows map[Handle]window
var messageQueue chan *Msg
var messageDispatcher func(msg *Msg) any

var ErrMessageQueueFull = errors.New("message queue is full")
var ErrInvalidWindow = errors.New("invalid window handle")

func PostMessage(win Handle, msg Message) error {
	if win != nil {
		w, ok := windows[win]
		if !ok {
			return ErrInvalidWindow
		}
		if _, ok := msg.(InputMessage); ok && !w.Enabled() {
			return nil
		}
	}

	select {
	case messageQueue <- &Msg{Window: win, Message: msg}:
	default:
		return ErrMessageQueueFull
	}
	return nil
}

func SendMessage(win Handle, msg Message) (any, error) {
	w, ok := windows[win]
	if !ok {
		return nil, ErrInvalidWindow
	}
	if _, ok := msg.(InputMessage); ok && !w.Enabled() {
		return nil, nil
	}

	return w.CallWndProc(&Msg{Window: win, Message: msg}), nil
}

func Run(f func()) int {
	windows = make(map[Handle]window)
	messageQueue = make(chan *Msg, msgQueueSize)
	messageDispatcher = func(msg *Msg) any {
		win, ok := windows[msg.Window]
		if !ok {
			return nil
		}
		return win.CallWndProc(msg)
	}

	defer func() {
		// destroy windows
		for _, h := range slices.Collect(maps.Keys(windows)) {
			// Child windows may have been destroyed already by its parent.
			if win, ok := windows[h]; ok {
				// Destroy all top-level windows.
				// Child windows will be destroyed automatically.
				topLevel, ok := win.(*topLevelWindow)
				if !ok {
					continue
				}
				if mu := topLevel.Menu(); mu != nil {
					chkerr.MustOK(DestroyMenu(mu))
				}
				chkerr.MustOK(DestroyWindow(h))

			}
		}
		windows = nil
		close(messageQueue)
		messageDispatcher = nil
	}()

	f()

	for msg := range messageQueue {
		if msg.Window == nil {
			switch msg := msg.Message.(type) {
			case MsgQuit:
				return int(msg)
			case MsgPostFunc:
				msg()
			}
			continue
		}
		messageDispatcher(msg)
	}
	return 0
}

// SetMessageDispatcher sets the message dispatcher that
func SetMessageDispatcher(dispatcher func(msg *Msg, prev func(*Msg) any) any) {
	prev := messageDispatcher
	messageDispatcher = func(msg *Msg) any {
		return dispatcher(msg, prev)
	}
}

func Post(f func()) error {
	return PostMessage(nil, MsgPostFunc(f))
}

func PostQuitMessage(exitCode int) error {
	return PostMessage(nil, MsgQuit(exitCode))
}

func ClientToScreen(win Handle, x, y metrics.DP) (screenX, screenY metrics.DP, err error) {
	screenX, screenY = x, y
	for win != nil {
		w, ok := windows[win]
		if !ok {
			err = ErrInvalidWindow
			return
		}
		rect := w.Rect()
		screenX += rect.Left
		screenY += rect.Top
		win = w.Parent()
	}
	return
}

func ScreenToClient(win Handle, x, y metrics.DP) (clientX, clientY metrics.DP, err error) {
	clientX, clientY = x, y
	for win != nil {
		w, ok := windows[win]
		if !ok {
			err = ErrInvalidWindow
			return
		}
		rect := w.Rect()
		clientX -= rect.Left
		clientY -= rect.Top
		win = w.Parent()
	}
	return
}

func NewWindow(title string, size metrics.Size) (h Handle, err error) {
	h = newHandle()
	var win = NewTopLevelWindow(title, &metrics.Rect{
		Left:   defaultWindowLeftTop.X,
		Top:    defaultWindowLeftTop.Y,
		Right:  defaultWindowLeftTop.X + size.Width,
		Bottom: defaultWindowLeftTop.Y + size.Height,
	})
	defaultWindowLeftTop.X += metrics.DP(30)
	defaultWindowLeftTop.Y += metrics.DP(30)
	if defaultWindowLeftTop.X > 1000 || defaultWindowLeftTop.Y > 1000 {
		defaultWindowLeftTop.X, defaultWindowLeftTop.Y = 0, 0
	}
	windows[h] = win
	return h, nil
}

func CloseWindow(win Handle) (err error) {
	allow, err := SendMessage(win, MsgClosing{})
	if err != nil {
		return
	}
	if !allow.(bool) {
		return
	}
	if _, err = SendMessage(win, MsgClosed{}); err != nil {
		return
	}
	return DestroyWindow(win)
}

func InvalidateWindow(win Handle, rect *metrics.Rect) (err error) {
	_, err = SendMessage(win, MsgPaint(*rect))
	return err
}

func DestroyWindow(win Handle) (err error) {
	if _, err = SendMessage(win, MsgDestroying{}); err != nil {
		return
	}
	if _, err = SendMessage(win, MsgDestroyed{}); err != nil {
		return
	}
	delete(windows, win)
	return nil
}

func SetOnSizeChangedListener(win Handle, listener func(size metrics.Size)) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	w.SetOnSizeChangedListener(listener)
	return nil
}

func SetOnDestroyListener(win Handle, listener func()) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	w.SetOnDestroyListener(win, listener)
	return nil
}

func SetOnCloseListener(win Handle, listener func() bool) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	w.(*topLevelWindow).SetOnCloseListener(listener)
	return nil
}

func ClientRect(win Handle) (rect metrics.Rect, err error) {
	w, ok := windows[win]
	if !ok {
		err = ErrInvalidWindow
		return
	}
	rect = w.Rect()
	width, height := rect.Width(), rect.Height()
	rect.Left = min(10, width)
	rect.Top = min(10, height)
	rect.Right = max(rect.Left, width-rect.Left-10)
	rect.Bottom = max(rect.Top, height-rect.Top-10)
	return
}

func WindowRect(win Handle) (rect metrics.Rect, err error) {
	w, ok := windows[win]
	if !ok {
		err = ErrInvalidWindow
		return
	}
	rect = w.Rect()
	return
}

func SetWindowRect(win Handle, rect metrics.Rect) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	w.SetRect(rect)
	return nil
}

func SetMenu(win Handle, m Handle) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	_, ok = menus[m]
	if !ok {
		return ErrInvalidMenu
	}
	w.(*topLevelWindow).SetMenu(m)
	return nil
}

func WindowMenu(win Handle) (Handle, error) {
	w, ok := windows[win]
	if !ok {
		return nil, ErrInvalidWindow
	}
	return w.(*topLevelWindow).Menu(), nil
}

func RefreshMenu(win Handle) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	return w.(*topLevelWindow).RefreshMenu()
}

func TrackPopupMenu(win Handle, menuToTrack Handle, spec *native.TrackPopupSpec) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	return w.(*topLevelWindow).TrackPopupMenu(menuToTrack, spec)
}

func EnableDrawDebugRect(win Handle, rects func() iter.Seq[native.DebugRect]) (layer Handle, err error) {
	return NewPanel(win)
}

func AddChild(parent, child Handle) error {
	p, ok := windows[parent]
	if !ok {
		return ErrInvalidWindow
	}
	c, ok := windows[child]
	if !ok {
		return ErrInvalidWindow
	}
	if oldParentHandle := windows[child].Parent(); oldParentHandle != nil {
		oldParent, ok := windows[oldParentHandle]
		if !ok {
			return ErrInvalidWindow
		}
		oldParent.RemoveChild(child)
	}
	p.appendChildSlice(child)
	c.setParentField(parent)
	return nil
}

func SetButtonOnClickListener(btn Handle, onClick func()) error {
	b, ok := windows[btn]
	if !ok {
		return ErrInvalidWindow
	}
	b.(*button).SetOnClickListener(onClick)
	return nil
}

// Debug_SimulateButtonClick simulates a button click for testing purposes.
func Debug_SimulateButtonClick(btn Handle) (err error) {
	_, err = SendMessage(btn, MsgMouseLeftDown{})
	return
}

func SetText(win Handle, text string) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	w.SetText(text)
	return nil
}

func Text(win Handle) (string, error) {
	w, ok := windows[win]
	if !ok {
		return "", ErrInvalidWindow
	}
	return w.Text(), nil
}

func ButtonMinimumSize(btn Handle, label string) (size metrics.Size, err error) {
	b, ok := windows[btn]
	if !ok {
		return size, ErrInvalidWindow
	}
	return b.(*button).MinimumSize(), nil
}

func SetLabelMultiline(lbl Handle, multiline bool) error {
	l, ok := windows[lbl]
	if !ok {
		return ErrInvalidWindow
	}
	l.(*label).SetMultiline(multiline)
	return nil
}

func SetLabelTextAlignment(lbl Handle, alignment native.TextAlignment) error {
	l, ok := windows[lbl]
	if !ok {
		return ErrInvalidWindow
	}
	l.(*label).SetTextAlignment(alignment)
	return nil
}

func SetLabelBackgroundColor(lbl Handle, color *color.NRGBA) error {
	l, ok := windows[lbl]
	if !ok {
		return ErrInvalidWindow
	}
	l.(*label).SetBackgroundColor(color)
	return nil
}

func SetControlDimensions(ctrl Handle, rect metrics.Rect) error {
	return SetWindowRect(ctrl, rect)
}

func SetControlEnabled(ctrl Handle, enabled bool) error {
	w, ok := windows[ctrl]
	if !ok {
		return ErrInvalidWindow
	}
	w.SetEnabled(enabled)
	return nil
}

func ControlEnabled(ctrl Handle) (bool, error) {
	w, ok := windows[ctrl]
	if !ok {
		return false, ErrInvalidWindow
	}
	return w.Enabled(), nil
}

func TextDrawingSize(control Handle, text string, multiline bool, maxWidth metrics.DP) (size metrics.Size, err error) {
	_, ok := windows[control]
	if !ok {
		err = ErrInvalidWindow
		return
	}
	const charWidth = metrics.DP(8)
	const charHeight = metrics.DP(16)
	var lines = 0
	var width metrics.DP
	for _, ch := range text {
		if ch == '\n' {
			if multiline {
				lines++
			}
			continue
		}
		if w := width + charWidth; maxWidth > 0 && w > maxWidth {
			lines++
			width = charWidth
		} else {
			width = w
		}
	}
	if len(text) > 0 {
		lines++
	}
	return metrics.Size{
		Width:  width,
		Height: metrics.DP(lines) * charHeight,
	}, nil
}

func SetPanelBackgroundColor(handle Handle, color *color.NRGBA) error {
	p, ok := windows[handle]
	if !ok {
		return ErrInvalidWindow
	}
	p.(*panel).SetBackgroundColor(color)
	return nil
}

func SetWndProc(win Handle, wndProc WndProc) error {
	w, ok := windows[win]
	if !ok {
		return ErrInvalidWindow
	}
	w.SetWndProc(wndProc)
	return nil
}
