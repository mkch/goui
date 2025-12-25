package goui

import (
	"fmt"
	"iter"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/internal/tricks"
	"github.com/mkch/goui/native"
)

type Context struct {
	app    *App    // can't be nil
	window *window // can't be nil
}

// newMockContext creates and returns a new mock goui.Context for testing.
func newMockContext(config *AppConfig) *Context {
	return &Context{
		app:    NewApp(config),
		window: &window{},
	}
}

// NativeWindow returns the native window handle associated with this context.
func (ctx *Context) NativeWindow() native.Handle {
	return ctx.window.Handle
}

// MessageBox shows a message box with the given title, message and icon
// associated with this context's window.
func (ctx *Context) MessageBox(title, message string, icon MessageBoxIcon) {
	native.MessageBox(ctx.NativeWindow(), title, message, native.MessageBoxIcon(icon))
}

type Widget interface {
	WidgetID() ID
	CreateElement(ctx *Context) (Element, error)
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
	// Debug can be nil, in which case no debug features are enabled.
	Debug *Debug
}

// Debug is the debug configuration for the app.
type Debug struct {
	// Layout debugging features.
	// Nil or a pointer to false value means disabled, a pointer to true means enabled.
	Layout *bool
}

// NewApp creates and returns a new App instance.
// The app is setup with the given config. If config is nil, default configuration is used.
func NewApp(config *AppConfig) *App {
	return &App{
		debug:   gg.If(config == nil, nil, (*tricks.Debug)(config.Debug).Clone()),
		app:     native.NewApp(),
		windows: make(map[ID]*window),
	}
}

func (app *App) Run() int {
	ctx := &Context{app: app}
	for _, window := range app.windows {
		if window.Window.Root != nil {
			ctx.window = window
			elem, layouter, err := buildElementTree(ctx, window.Window.Root)
			if err != nil {
				errortrace.Panic(err)
			}
			window.Root = elem
			window.Layouter = layouter
			if err := layoutWindow(&Context{app, window}); err != nil {
				errortrace.Panic(err)
			}
		}
	}

	return app.app.Run()
}

func (app *App) Exit(exitCode int) {
	app.app.Quit(exitCode)
}

type MessageBoxIcon native.MessageBoxIcon

const (
	MessageBoxNone        = MessageBoxIcon(native.MessageBoxNone)
	MessageBoxIconInfo    = MessageBoxIcon(native.MessageBoxIconInfo)
	MessageBoxIconWarning = MessageBoxIcon(native.MessageBoxIconWarning)
	MessageBoxIconError   = MessageBoxIcon(native.MessageBoxIconError)
)

// MessageBox shows a message box with the given title, message and icon.
func MessageBox(title, message string, icon MessageBoxIcon) {
	native.MessageBox(nil, title, message, native.MessageBoxIcon(icon))
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

func performLayoutWindow(ctx *Context, width, height int) (err error) {
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
		config.ID = ValueID(&config) // unique key is required to insert into the map
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
	native.SetWindowOnSizeChangedListener(handle, func(width, height int) {
		if err := performLayoutWindow(&Context{app, window}, width, height); err != nil {
			panic(err)
		}
	})
	native.SetWindowOnCloseListener(handle, config.OnClose)
	if app.debug.LayoutDebugEnabled() {
		native.EnableDrawDebugRect(handle, func() iter.Seq[native.DebugRect] {
			if window.Layouter == nil {
				return func(yield func(native.DebugRect) bool) {}
			}
			return allLayouterDebugOutlines(window.Layouter)
		})
	}
	app.windows[config.ID] = window
	return nil
}
