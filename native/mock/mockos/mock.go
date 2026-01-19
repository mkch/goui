package mockos

import (
	"slices"
	"unsafe"

	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui/metrics"
)

// Handle represents a handle of a mock OS element.
type Handle unsafe.Pointer

// newHandle creates a new unique Handle.
// If a Handle is not in use, it may be reused.
func newHandle() Handle {
	return Handle(new(int))
}

// Msg is a window message with its target window.
type Msg struct {
	Window  Handle
	Message Message
}

// WndProc is a window procedure function.
// The prev function can be called to invoke the previous window procedure.
type WndProc func(msg *Msg, prev func(*Msg) any) any

// window is the base interface for mock OS window elements.
type window interface {
	// Parent returns the parent of this window.
	// For now only non-[topLevelWindow] have a non-nil parent.
	Parent() Handle
	// Children returns a copy of the children slice.
	Children() []Handle
	// RemoveChild removes a child from its parent.
	RemoveChild(child Handle) error
	// Text returns the text of this window.
	Text() string
	// SetText sets the text of this window.
	SetText(text string)
	// SetWndProc sets the window procedure of this window.
	SetWndProc(handle WndProc)
	// CallWndProc calls the window procedure of this window.
	CallWndProc(*Msg) any
	// Rect returns the rectangle of this window in its parent's client coordinates.
	// If this window is a [topLevelWindow], the coordinates are in screen coordinates.
	Rect() metrics.Rect
	// SetRect sets the position and size of this window in its parent's client coordinates.
	// If this window is a [topLevelWindow], the coordinates are in screen coordinates.
	SetRect(rect metrics.Rect)
	// SetOnSizeChangedListener sets a listener function that is called when the size of this window changes.
	SetOnSizeChangedListener(onSizeChanged func(size metrics.Size))
	// SetOnDestroyListener sets a listener function that is called when this window is destroyed.
	SetOnDestroyListener(handle Handle, listener func())
	// SetEnabled sets whether this window is enabled.
	// A disabled window does not receive input events.
	SetEnabled(enabled bool)
	// Enabled returns whether this window is enabled.
	Enabled() bool

	// setParentField sets the parent field directly without modifying the parent's children slice.
	setParentField(parent Handle)
	// appendChildSlice appends a child to the children slice directly without modifying the child's parent field.
	appendChildSlice(child Handle)
}

// abstractWindow implements [window].
// It can be embedded in other window types to provide common functionality.
type abstractWindow struct {
	parent         Handle
	children       []Handle
	text           string
	wndProc        func(msg *Msg) any
	rect           metrics.Rect
	disabled       bool
	onSizedChanged func(size metrics.Size)
	onDestroy      func()
}

// initAbstractWindow initializes the abstractWindow fields and returns parameter win.
// Embedders should call this in their constructor.
func initAbstractWindow(win *abstractWindow, rect *metrics.Rect) *abstractWindow {
	win.rect = *rect
	win.wndProc = func(msg *Msg) any {
		switch msg.Message.(type) {
		case MsgSizedChanged:
			win.callOnSizeChangedListener()
		case MsgDestroying:
			for _, child := range win.children {
				chkerr.MustOK(DestroyWindow(child))
			}
		case MsgDestroyed:
			win.callOnDestroyListener()
		}
		return nil
	}
	return win
}

func (win *abstractWindow) callOnSizeChangedListener() {
	if win.onSizedChanged != nil {
		win.onSizedChanged(win.rect.Size())
	}
}

func (win *abstractWindow) callOnDestroyListener() {
	if win.onDestroy != nil {
		win.onDestroy()
	}
}

func (win *abstractWindow) Parent() Handle {
	return win.parent
}

func (win *abstractWindow) setParentField(parent Handle) {
	win.parent = parent
}

func (win *abstractWindow) appendChildSlice(child Handle) {
	win.children = append(win.children, child)
}

func (win *abstractWindow) Children() []Handle {
	return slices.Clone(win.children)
}

func (win *abstractWindow) RemoveChild(child Handle) error {
	c, ok := windows[child]
	if !ok {
		return ErrInvalidWindow
	}
	c.setParentField(nil)
	win.children = slices.DeleteFunc(win.children, func(h Handle) bool { return h == child })
	return nil
}

func (win *abstractWindow) Text() string {
	return win.text
}

func (win *abstractWindow) SetText(text string) {
	win.text = text
}

func (win *abstractWindow) SetWndProc(proc WndProc) {
	prev := win.wndProc
	win.wndProc = func(msg *Msg) any {
		return proc(msg, prev)
	}
}

func (win *abstractWindow) CallWndProc(msg *Msg) any {
	return win.wndProc(msg)
}

func (win *abstractWindow) Rect() metrics.Rect {
	return win.rect
}

func (win *abstractWindow) SetRect(rect metrics.Rect) {
	oldWidth := win.rect.Width()
	oldHeight := win.rect.Height()
	win.rect = rect
	if win.rect.Width() != oldWidth || win.rect.Height() != oldHeight {
		win.callOnSizeChangedListener()
	}
}

func (win *abstractWindow) SetOnSizeChangedListener(onSizeChanged func(size metrics.Size)) {
	win.onSizedChanged = onSizeChanged
}

func (win *abstractWindow) SetOnDestroyListener(handle Handle, listener func()) {
	win.onDestroy = listener
}

func (win *abstractWindow) SetEnabled(enabled bool) {
	win.disabled = !enabled
}

func (win *abstractWindow) Enabled() bool {
	return !win.disabled
}
