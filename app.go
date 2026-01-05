package goui

import (
	"errors"
	"fmt"
	"iter"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/internal/tricks"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type Context struct {
	app    *App    // can't be nil
	window *window // can't be nil
}

// newMockContext creates and returns a new mock goui.Context for testing.
func newMockContext(config *AppConfig) (ctx *Context) {
	ctx = &Context{
		app: &App{
			debug: configDebug(config),
		},
		window: &window{},
	}
	return
}

// NativeWindow returns the native window handle associated with this context.
func (ctx *Context) NativeWindow() native.Handle {
	return ctx.window.Handle
}

func (ctx *Context) App() *App {
	return ctx.app
}

type Widget interface {
	WidgetID() ID
	CreateElement(ctx *Context, parent Element) (Element, error)
}

type Container interface {
	Widget
	NumChildren() int
	Child(n int) Widget
	// Exclusive is a marker method to distinguish StatefulWidget, StatelessWidget and Container.
	Exclusive(Container)
}

// App is the main application object that manages the GUI application's lifecycle and windows.
// There should not be more than one App instance per goroutine.
type App struct {
	debug   *tricks.Debug
	app     native.App
	windows map[ID]*window
}

// Post posts a function to be executed on the main GUI goroutine.
func (app *App) Post(f func()) error {
	return app.app.Post(f)
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
}

// configDebug returns a cloned *tricks.Debug from the given AppConfig.
// If config is nil, it returns nil.
func configDebug(config *AppConfig) *tricks.Debug {
	return gg.IfFunc(config == nil,
		func() *tricks.Debug { return nil },
		func() *tricks.Debug { return (*tricks.Debug)(config.Debug).Clone() })
}

// NewApp creates and returns a new App instance.
// The app is setup with the given config. If config is nil, default configuration is used.
func NewApp(config *AppConfig) (app *App) {
	return &App{
		debug:   configDebug(config),
		app:     native.NewApp(),
		windows: make(map[ID]*window),
	}
}

func (app *App) Run() int {
	return app.app.Run()
}

func (app *App) Exit(exitCode int) {
	app.app.Quit(exitCode)
}

func layoutWindow(ctx *Context) error {
	_, _, width, height, err := native.WindowClientRect(ctx.window.Handle)
	if err != nil {
		return err
	}
	if err := performLayoutWindow(ctx, width, height); err != nil {
		return err
	}
	return nil
}

func performLayoutWindow(ctx *Context, width, height metrics.DP) (err error) {
	if ctx.window.Layouter == nil {
		return nil
	}
	_, err = ctx.window.Layouter.Layout(ctx, Constraints{
		MinWidth:  0,
		MinHeight: 0,
		MaxWidth:  width,
		MaxHeight: height,
	})
	if err != nil {
		return err
	}
	return ctx.window.Layouter.PositionAt(0, 0)

}

// ErrInvalidWindowRoot is returned when creating a window with an invalid root widget,
// such as a Menu or MenuItem.
var ErrInvalidWindowRoot = errors.New("invalid window root")

// CreateWindow creates a new window with the given configuration.
// If config is nil, a default configuration is used.
// If a window with the same ID already exists, it returns an error.
func (app *App) CreateWindow(config *Window) error {
	if config == nil {
		config = &Window{
			Width: 800, Height: 600,
		}
	}
	if config.ID == nil {
		config.ID = UniqueID() // unique key is required to insert into the map
	}
	if app.windows[config.ID] != nil {
		return fmt.Errorf("window with ID %v already exists", config.ID)
	}
	handle, err := native.CreateWindow(config.Title, config.Width, config.Height)
	if err != nil {
		return err
	}
	window := &window{
		ID:     config.ID,
		Handle: handle,
	}
	ctx := &Context{app, window}
	native.SetWindowOnSizeChangedListener(handle, func(width, height metrics.DP) {
		if err := performLayoutWindow(ctx, width, height); err != nil {
			errortrace.Panic(err)
		}
	})
	native.SetWindowOnCloseListener(handle, func() bool {
		if config.OnClose == nil {
			return true
		}
		return config.OnClose(ctx)
	})
	native.SetWindowOnDestroyListener(handle, func() {
		if config.OnDestroy != nil {
			config.OnDestroy(ctx)
		}
		delete(app.windows, config.ID)
	})
	if app.debug.LayoutOutlineEnabled() {
		layer, err := native.EnableDrawDebugRect(handle, func() iter.Seq[native.DebugRect] {
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
		elem, layouter, err := buildElementTree(ctx, config.Root)
		if err != nil {
			errortrace.Panic(err)
		}
		if containsNativeMenu(elem) {
			// Menu does not attach to window may cause resource leak.
			return ErrInvalidWindowRoot
		}
		window.Root = elem
		window.Layouter = layouter
		if err := layoutWindow(ctx); err != nil {
			errortrace.Panic(err)
		}
	}

	if config.Menu != nil {
		elem, _, err := buildElementTree(ctx, config.Menu)
		if err != nil {
			errortrace.Panic(err)
		}
		if nm := unwrapNativeMenu(elem); nm != nil {
			window.Menu = elem
			err = native.SetWindowMenu(window.Handle, nm)
			if err != nil {
				errortrace.Panic(err)
			}
		}
	}

	app.windows[config.ID] = window
	return nil
}

// unwrapNativeMenu unwraps the given Element to find the nearest underlying [NativeMenuElement].
// Any wrapper that is not [StatelessWidget] or [StatefulWidget] are skipped.
func unwrapNativeMenu(element Element) native.Handle {
	h, found := LookupChild(element, func(e Element) (h native.Handle, found bool, cont bool) {
		if nativeMenu, found := e.(NativeMenuElement); found {
			h = nativeMenu.NativeMenu()
			return h, true, false
		}
		if _, isStateless := e.Widget().(StatelessWidget); isStateless {
			cont = true
		} else if _, isStateful := e.Widget().(StatefulWidget); isStateful {
			cont = true
		}
		return
	})
	if found {
		return h
	}
	return nil
}

// containsNativeMenu reports whether the given element tree contains any [NativeMenuElement].
func containsNativeMenu(element Element) bool {
	_, found := LookupChild(element, func(e Element) (_ struct{}, found bool, cont bool) {
		_, found = e.(NativeMenuElement)
		if found {
			return
		}
		cont = true
		return
	})
	return found
}

var ErrNoSuchWindow = errors.New("no such window exists")

// CloseWindow closes the window with the given ID.
// If no such window exists, it returns [ErrNoSuchWindow].
func (app *App) CloseWindow(windowID ID) error {
	win := app.windows[windowID]
	if win == nil {
		return ErrNoSuchWindow
	}
	return native.CloseWindow(app.windows[windowID].Handle)
}
