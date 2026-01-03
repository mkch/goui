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

func (app *App) CreateWindow(config Window) error {
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
		Window: config,
		ID:     config.ID,
		Handle: handle,
	}
	ctx := &Context{app, window}
	native.SetWindowOnSizeChangedListener(handle, func(width, height metrics.DP) {
		if err := performLayoutWindow(ctx, width, height); err != nil {
			panic(err)
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

	if window.Window.Root != nil {
		ctx.window = window
		elem, layouter, err := buildElementTree(ctx, window.Window.Root)
		if err != nil {
			errortrace.Panic(err)
		}
		window.Root = elem
		window.Layouter = layouter
		if err := layoutWindow(ctx); err != nil {
			errortrace.Panic(err)
		}
	}

	app.windows[config.ID] = window
	return nil
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
