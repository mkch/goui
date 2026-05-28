package native

import (
	"image/color"
	"iter"

	"github.com/mkch/goui/metrics"
)

// Handle represents a platform-specific GUI object.
type Handle any

type TextAlignment int

const (
	Left TextAlignment = iota
	Right
	Center
)

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

type TrackPopupSpec struct {
	X, Y metrics.DP // Screen coordinates.
}

type MessageBoxIcon int

const (
	MessageBoxIconNone MessageBoxIcon = iota
	MessageBoxIconInfo
	MessageBoxIconWarning
	MessageBoxIconQuestion
	MessageBoxIconError
)

type MessageBoxButton int

const (
	MessageBoxButtonOK MessageBoxButton = iota
	MessageBoxButtonYesNo
)

type MessageBoxReturn int

const (
	MessageBoxReturnCancel MessageBoxReturn = iota
	MessageBoxReturnOK
	MessageBoxReturnYes
	MessageBoxReturnNo
)

type DebugRect struct {
	Left, Top, Right, Bottom metrics.DP
	Highlight                bool
}

type OS interface {
	App_Run(initialize func()) int
	App_Post(f func()) error
	App_Quit(exitCode int) error
	App_Destroy()
	// App_AddMouseEventListener adds a mouse event listener to all windows in the application.
	// It returns a function to remove the listener.
	// Parameter win is the window whose client coordinates are used for the event positions.
	App_AddMouseEventListener(win Handle, listener MouseEventListener) (remove func())

	// NewWindow creates a native window with the specified configuration.
	NewWindow(title string, size metrics.Size) (h Handle, err error)
	Window_Close(win Handle) error
	Window_Invalidate(win Handle, rect *metrics.Rect) error
	Window_Destroy(win Handle) error
	Window_SetOnSizeChangedListener(handle Handle, onSizeChanged func(size metrics.Size)) error
	Window_SetOnDestroyListener(handle Handle, listener func()) error
	Window_SetOnCloseListener(handle Handle, listener func() bool) error
	Window_ClientRect(handle Handle) (rect metrics.Rect, err error)
	Window_Menu(win Handle) (Handle, error)
	Window_SetMenu(win Handle, m Handle) error
	Window_RefreshMenu(win Handle) error
	Window_TrackPopupMenu(win Handle, menuToTrack Handle, spec *TrackPopupSpec) error
	Window_EnableDrawDebugRect(win Handle, rects func() iter.Seq[DebugRect]) (layer Handle, err error)
	Window_Enabled(win Handle) (bool, error)
	Window_SetEnabled(win Handle, enabled bool) error

	NewMenu(popup bool) (Handle, error)
	Menu_Destroy(m Handle) error

	NewMenuItem(parent Handle, title string, separator bool) (handle Handle, err error)
	MenuItem_Destroy(item Handle) error
	MenuItem_SetTitle(item Handle, title string) error
	MenuItem_SetDisabled(item Handle, disabled bool) error
	MenuItem_SetSubmenu(item Handle, submenu Handle) error
	MenuItem_SetOnClickListener(item Handle, listener func())

	NewButton(parent Handle, title string) (handle Handle, err error)
	Button_SetOnClickListener(btn Handle, onClick func()) error
	Button_SetLabel(btn Handle, label string) error
	Button_MinimumSize(btn Handle, label string) (size metrics.Size, err error)

	NewLabel(parent Handle, title string) (handle Handle, err error)
	Label_SetMultiline(lbl Handle, multiline bool) error
	Label_SetTextAlignment(lbl Handle, alignment TextAlignment) error
	Label_SetText(handle Handle, text string) error
	// Label_SetBackgroundColor sets the background color of the label.
	// If color is nil, the default color is used.
	Label_SetBackgroundColor(lbl Handle, color *color.NRGBA) error

	NewTextField(parent Handle, initialValue string, password bool) (handle Handle, err error)
	TextField_Text(tf Handle) (text string, err error)
	TextField_SetText(tf Handle, text string) error

	Control_SetDimensions(ctrl Handle, rect metrics.Rect) error
	Control_SetEnabled(ctrl Handle, enabled bool) error
	Control_Enabled(ctrl Handle) (bool, error)
	Control_SetOnDestroyListener(handle Handle, onClose func()) error
	Control_Destroy(win Handle) error

	// Control_TextDrawingSize returns the size required to draw the specified text
	// in the given control.
	// If multiline is true, the line ending characters are considered as line breaks.
	// If maxWidth is greater than zero, the returned width will not exceed maxWidth.
	Control_TextDrawingSize(control Handle, text string, multiline bool, maxWidth metrics.DP) (size metrics.Size, err error)

	NewPanel(parent Handle) (handle Handle, err error)
	Panel_SetBackgroundColor(handle Handle, color *color.NRGBA) error

	Util_ClientCoordinatesConv(from, to Handle, x, y metrics.DP) (newX, newY metrics.DP, err error)
	// ClientToScreenPt converts the point (x, y) in the client area of the specified window
	// to screen coordinates.
	Util_ClientToScreen(win Handle, x, y metrics.DP) (screenX, screenY metrics.DP, err error)

	MessageBox(parent Handle, title, message string, icon MessageBoxIcon, button MessageBoxButton) (ret MessageBoxReturn, err error)
}
