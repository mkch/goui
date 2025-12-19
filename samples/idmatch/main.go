package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
)

var app = goui.NewApp()

func main() {
	app.CreateWindow(goui.Window{
		DebugLayout: true,
		Title:       "goui idmatch demo",
		Width:       600,
		Height:      400,
		Root:        Root(),
	})
	app.Run()
}

type Person struct {
	ID   int
	Name string
	Age  int
}

var personList = []Person{
	{0, "Alice", 30},
	{1, "Bob", 25},
	{2, "Charlie", 35},
}

func Root() goui.StatefulWidget {
	return goui.StatefulWidgetFunc(func(ctx *goui.Context) (state *goui.WidgetState) {
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
								OnClick: func() {
									// Sort personList by Name
									slices.SortStableFunc(personList, func(a, b Person) int {
										return strings.Compare(a.Name, b.Name)
									})
									// Update the whole Root widget to rebuild children
									state.Update(func() {})
								},
							},
						},

						&widgets.Padding{
							Top: 20,
							Widget: &widgets.Button{
								Label: "Sort by age",
								OnClick: func() {
									// Sort personList by Age
									slices.SortStableFunc(personList, func(a, b Person) int {
										return a.Age - b.Age
									})
									// Update the whole Root widget to rebuild children
									state.Update(func() {})
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
		func(ctx *goui.Context) (state *goui.WidgetState) {
			return &goui.WidgetState{
				Build: func() goui.Widget {
					return &widgets.Button{
						ID:    goui.ValueID(p.ID),
						Label: fmt.Sprintf("%s (%d years old) - Clicked %d times", p.Name, p.Age, clicked),
						OnClick: func() {
							state.Update(func() {
								clicked++
							})
						},
					}
				},
			}
		})

}
