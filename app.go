package goui

import (
	"fmt"
	"iter"

	"github.com/mkch/goui/native"
)

type Context struct {
	app    *App    // can't be nil
	window *window // can't be nil
}

// newMockContext creates and returns a new mock goui.Context for testing.
func newMockContext() *Context {
	return &Context{
		window: &window{},
	}
}

// NativeWindow returns the native window handle associated with this context.
func (ctx *Context) NativeWindow() native.Handle {
	return ctx.window.Handle
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
	app     native.App
	windows map[ID]*window
}

// Post posts a function to be executed on the main GUI goroutine.
func (app *App) Post(f func()) error {
	return app.app.Post(f)
}

func NewApp() *App {
	return &App{
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
				panic(err)
			}
			window.Root = elem
			window.Layouter = layouter
			if err := layoutWindow(&Context{app, window}); err != nil {
				panic(err)
			}
		}
	}

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
	if config.DebugLayout {
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
