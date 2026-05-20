package goui

import (
	"errors"
	"fmt"
	"iter"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/internal/util"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
	"github.com/mkch/goui/native/mock"
)

var appOS native.OS
var appDebug *Debug
var appWindows map[ID]*window

// OS returns the native OS interface used by the application.
func OS() native.OS {
	return appOS
}

// Post posts a function to be executed on the main GUI goroutine.
func Post(f func()) error {
	return appOS.App_Post(f)
}

// AppConfig is the configuration for creating a new App.
type AppConfig struct {
	// Debug is the debug configuration for the app.
	// If Debug is non-nil, debug mode is on, and the fields in Debug control which features are enabled.
	// If Debug is nil, debug mode is off.
	Debug *Debug
}

// Debug is the debug configuration for the app.
type Debug struct {
	// LayoutOutline specifies whether layout outlines are drawn in debug mode.
	LayoutOutline bool
	// If non-nil, the native.Handle of each [Element] whose ID is non-nil will be recorded in this map.
	NativeHandleRecords map[ID]native.Handle
}

// layoutOutlineEnabled returns whether layout outline is enabled in debug mode.
// If debug is nil, it returns false.
func (debug *Debug) layoutOutlineEnabled() bool {
	return debug != nil && debug.LayoutOutline
}

// recordNativeHandle records the native handle(if any) of the element in debug mode.
// If debug is nil, or debug.NativeHandleRecords is nil, it does nothing.
func (debug *Debug) recordNativeHandle(element Element) {
	if debug == nil || debug.NativeHandleRecords == nil {
		return
	}
	id := element.Widget().WidgetID()
	if id == nil {
		return
	}
	if ctrl, ok := element.(ControlElement); ok {
		appDebug.NativeHandleRecords[id] = ctrl.NativeControl()
	} else if mu, ok := element.(NativeMenuElement); ok {
		appDebug.NativeHandleRecords[id] = mu.NativeMenu()
	} else if mi, ok := element.(NativeMenuItemElement); ok {
		appDebug.NativeHandleRecords[id] = mi.NativeMenuItem()
	}
}

// deleteNativeHandleRecord deletes the native handle record of the element in debug mode.
// If debug is nil, or debug.NativeHandleRecords is nil, it does nothing.
func (debug *Debug) deleteNativeHandleRecord(element Element) {
	if debug == nil || debug.NativeHandleRecords == nil {
		return
	}
	id := element.Widget().WidgetID()
	if id == nil {
		return
	}
	delete(debug.NativeHandleRecords, id)
}

func (debug *Debug) clone() (result *Debug) {
	if debug == nil {
		return nil
	}
	result = &Debug{}
	*result = *debug
	return
}

// configDebug returns a cloned *Debug from the given AppConfig.
// If config is nil, it returns nil.
func configDebug(config *AppConfig) *Debug {
	return gg.IfFunc(config == nil,
		func() *Debug { return nil },
		func() *Debug { return config.Debug.clone() })
}

// Run does the following things sequentially:
//   - initializes the application
//   - calls f
//   - runs the main event loop
//
// It returns the exit code passed to [Exit].
// No other goui functions should be called before or after Run.
func Run(f func(), config *AppConfig) int {
	return runOS(newOS(), f, config)
}

// RunAndExit calls [os.Exit]([Run](f, config)).
// It is convenience for applications that want to exit with the exit code returned by [Run].
func RunAndExit(f func(), config *AppConfig) {
	os.Exit(Run(f, config))
}

// runOS runs the application with the given native OS implementation.
func runOS(os native.OS, f func(), config *AppConfig) int {
	defer func() {
		appOS = nil
		appDebug = nil
		for _, w := range appWindows {
			if err := appOS.Window_Destroy(w.Handle); err != nil {
				errortrace.Panic(err)
			}
		}
		appWindows = nil
	}()
	appOS = os
	appDebug = configDebug(config)
	appWindows = make(map[ID]*window)
	return appOS.App_Run(f)
}

// runContext initializes the application and creates a Context with a new window.
func runContext(os native.OS, f func(ctx *Context), config *AppConfig) int {
	defer func() {
		appOS = nil
		appDebug = nil
		for _, w := range appWindows {
			if err := appOS.Window_Destroy(w.Handle); err != nil {
				errortrace.Panic(err)
			}
		}
		appWindows = nil
	}()
	appOS = os
	appDebug = configDebug(config)
	appWindows = make(map[ID]*window)
	return appOS.App_Run(func() {
		id := UniqueID()
		CreateWindow(&Window{ID: id})
		f(&Context{appWindows[id]})
		Exit(0)
	})
}

// runContextMock calls [runContext] with a mock native OS implementation.
func runContextMock(f func(ctx *Context), config *AppConfig) int {
	return runContext(mock.NewOS(), f, config)
}

// Exit quits the main event loop of the application with the given exit code.
// The exit code will be returned by the [Run] function.
func Exit(exitCode int) {
	appOS.App_Quit(exitCode)
}

func layoutWindow(ctx *Context) error {
	rect, err := appOS.Window_ClientRect(ctx.window.Handle)
	if err != nil {
		return err
	}
	if err := performLayoutWindow(ctx, rect.Width(), rect.Height()); err != nil {
		return err
	}
	return nil
}

func performLayoutWindow(ctx *Context, width, height metrics.DP) (err error) {
	if ctx.window.Layouter == nil {
		return nil
	}
	_, err = ctx.window.Layouter.Layout(ctx, metrics.Constraints{
		MinWidth:  0,
		MinHeight: 0,
		MaxWidth:  width,
		MaxHeight: height,
	})
	if err != nil {
		return err
	}
	return ctx.window.Layouter.PositionAt(ctx, metrics.Point{X: 0, Y: 0})

}

// CreateWindow creates a new window with the given configuration.
// If config is nil, a default configuration is used.
// If a window with the same ID already exists, it returns an error.
func CreateWindow(config *Window) (err error) {
	if config == nil {
		config = &Window{
			Width: 800, Height: 600,
		}
	}
	if config.ID == nil {
		config.ID = UniqueID() // unique key is required to insert into the map
	}
	if appWindows[config.ID] != nil {
		return fmt.Errorf("window with ID %v already exists", config.ID)
	}

	handle, err := appOS.NewWindow(config.Title, metrics.Size{Width: config.Width, Height: config.Height})
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			appOS.Window_Destroy(handle)
		}
	}()

	window := &window{
		ID:     config.ID,
		Handle: handle,
	}
	ctx := &Context{window}
	appOS.Window_SetOnSizeChangedListener(handle, func(size metrics.Size) {
		if err = performLayoutWindow(ctx, size.Width, size.Height); err != nil {
			return
		}
	})
	appOS.Window_SetOnCloseListener(handle, func() bool {
		if config.OnClose == nil {
			return true
		}
		return config.OnClose(ctx)
	})
	appOS.Window_SetOnDestroyListener(handle, func() {
		if config.OnDestroy != nil {
			config.OnDestroy(ctx)
		}
		delete(appWindows, config.ID)
	})
	if appDebug.layoutOutlineEnabled() {
		layer, err := appOS.Window_EnableDrawDebugRect(handle, func() iter.Seq[native.DebugRect] {
			if window.Layouter == nil {
				return func(yield func(native.DebugRect) bool) {}
			}
			return allLayouterDebugOutlines(window.Layouter)
		})
		if err != nil {
			return err
		}
		window.DebugLayer = layer
	}

	if config.Root != nil {
		window.Root, window.Layouter, err = buildElementTree(ctx, config.Root)
		if err != nil {
			return
		}
		if err = layoutWindow(ctx); err != nil {
			return
		}
	}

	if config.Menu != nil {
		var elem Element
		elem, err = BuildElementTree(ctx, config.Menu)
		if err != nil {
			return
		}
		menuHandle := unwrapNativeMenu(elem)
		if menuHandle == nil {
			return
		}
		window.Menu = elem
		err = appOS.Window_SetMenu(window.Handle, menuHandle)
		if err != nil {
			return
		}
	}

	appWindows[config.ID] = window
	return nil
}

// unwrapNativeMenu unwraps the given Element to find the nearest underlying [NativeMenuElement].
// If such a NativeMenuElement is found, its native.Handle is returned.
// Any wrapper that is neither stateful menu nor stateless menu are skipped.
func unwrapNativeMenu(element Element) native.Handle {
	h, found := util.LookupChild(element, func(e Element) (native.Handle, bool) {
		if nativeMenu, ok := e.(NativeMenuElement); ok {
			return nativeMenu.NativeMenu(), true // Found a menu
		}
		widget := e.Widget()
		// More precisely, [menu.StatelessMenu] and [menu.StatefulMenu] are expected here,
		// but we use statelessMenu and statefulMenu to avoid import cycle.
		type statelessMenu interface {
			StatelessComponent
			ExclusiveType(marker.TypeMenu)
		}
		type statefulMenu interface {
			StatefulComponent
			ExclusiveType(marker.TypeMenu)
		}
		if _, isStateless := widget.(statelessMenu); isStateless {
			return nil, false // May contain a menu, continue searching
		}
		if _, isStateful := widget.(statefulMenu); isStateful {
			return nil, false // May contain a menu, continue searching
		}
		// Can't contain a menu. Found a `Not Found`(nil).
		// See the element creation logic in package menu.
		return nil, true
	})
	if found {
		return h
	}
	return nil
}

// ErrNoSuchWindow is returned when trying to access a window using
// an ID that does not exist.
var ErrNoSuchWindow = errors.New("no such window exists")

// CloseWindow closes the window with the given ID.
// If no such window exists, it returns [ErrNoSuchWindow].
func CloseWindow(windowID ID) error {
	win := appWindows[windowID]
	if win == nil {
		return ErrNoSuchWindow
	}
	return appOS.Window_Close(appWindows[windowID].Handle)
}
