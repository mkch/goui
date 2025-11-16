package goui

import (
	"fmt"

	"github.com/mkch/goui/native"
)

type Context struct {
	window *window // can't be nil
}

type Widget interface {
	WidgetID() ID
	CreateElement(ctx *Context) (Element, error)
}

type Container interface {
	Widget
	NumChildren() int
	Child(n int) Widget
}

type App struct {
	app     native.App
	windows map[ID]*window
}

func NewApp() *App {
	return &App{
		app:     native.NewApp(),
		windows: make(map[ID]*window),
	}
}

func (app *App) Run() int {
	for _, window := range app.windows {
		if window.Window.Root != nil {
			ctx := &Context{window: window}
			elem, layouter, err := buildElementTree(ctx, window.Window.Root, nil)
			if err != nil {
				panic(err)
			}
			window.Root = elem
			window.Layouter = layouter
			if err := layoutWindow(window); err != nil {
				panic(err)
			}
		}
	}

	return app.app.Run()
}

func layoutWindow(window *window) error {
	_, _, width, height, err := native.WindowClientRect(window.Handle)
	if err != nil {
		return err
	}
	if err := performLayoutWindow(window, width, height); err != nil {
		return err
	}
	return nil
}

func performLayoutWindow(window *window, width, height int) (err error) {
	if window.Layouter == nil {
		return nil
	}
	_, err = window.Layouter.Layout(&Context{window: window}, Constraints{
		MinWidth:  0,
		MinHeight: 0,
		MaxWidth:  width,
		MaxHeight: height,
	})
	if err != nil {
		return err
	}
	return window.Layouter.PositionAt(0, 0)

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
		if err := performLayoutWindow(window, width, height); err != nil {
			panic(err)
		}
	})
	app.windows[config.ID] = window
	return nil
}
