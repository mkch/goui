package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	app.CreateWindow(goui.Window{
		OnClose: func() { app.Exit(0) },
		Title:   "goui idmatch demo",
		Width:   600,
		Height:  400,
		Root:    Root(),
	})
	app.Run()
}

type Person struct {
	ID   int
	Name string
	Age  int
}

var personList = []Person{
	{0, "Charlie", 35},
	{1, "Alice", 30},
	{2, "Bob", 25},
}

func Root() goui.StatefulWidget {
	return goui.StatefulWidgetFunc(func(ctx *goui.Context, updateState goui.UpdateStateFunc) *goui.WidgetState {
		return &goui.WidgetState{
			Build: func() goui.Widget {
				return &widgets.Column{
					Widgets: []goui.Widget{

						&widgets.Button{
							Label: "--HEADER-- (ID changes on every build)",
							ID:    goui.ValueID(time.Now()),
						},

						NewPersonWidget(0),
						NewPersonWidget(1),
						NewPersonWidget(2),

						&widgets.Padding{
							Top: 20,
							Widget: &widgets.Button{
								Label: "Sort by name",
								OnClick: func(ctx *goui.Context) {
									// Update the whole Root widget to rebuild children
									gg.MustOK(updateState(func() {
										// Sort personList by Name
										slices.SortStableFunc(personList, func(a, b Person) int {
											return strings.Compare(a.Name, b.Name)
										})
									}))
								},
							},
						},

						&widgets.Padding{
							Top: 20,
							Widget: &widgets.Button{
								Label: "Sort by age",
								OnClick: func(ctx *goui.Context) {
									// Update the whole Root widget to rebuild children
									gg.MustOK(updateState(func() {
										// Sort personList by Age
										slices.SortStableFunc(personList, func(a, b Person) int {
											return a.Age - b.Age
										})
									}))
								},
							},
						},
					},
				}
			},
		}
	})
}

func NewPersonWidget(n int) goui.StatefulWidget {
	p := personList[n]
	var clicked = 0
	return goui.NewStatefulWidget(
		goui.ValueID(p.ID),
		func(ctx *goui.Context, updateState goui.UpdateStateFunc) *goui.WidgetState {
			return &goui.WidgetState{
				Build: func() goui.Widget {
					return &widgets.Button{
						ID:    goui.ValueID(p.ID),
						Label: fmt.Sprintf("%s (%d years old) - Clicked %d times", p.Name, p.Age, clicked),
						OnClick: func(ctx *goui.Context) {
							gg.MustOK(updateState(func() { clicked++ }))
						},
					}
				},
			}
		})

}
