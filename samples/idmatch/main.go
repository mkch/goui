package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"slices"
	"strings"

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
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "goui idmatch demo",
		Width:     1200,
		Height:    800,
		Root:      Root(),
	})
	app.Run()
}

type Person struct {
	ID   int
	Name string
	Age  int
}

func (p *Person) String() string {
	return fmt.Sprintf("Person{ID:%d, Name:%s, Age:%d}", p.ID, p.Name, p.Age)
}

type State struct {
	goui.StateUpdater
	goui.NopDestroyer
	PersonList []Person
}

func NewState(ctx *goui.StateContext) *State {
	return &State{
		StateUpdater: goui.NewStateUpdater(ctx),
		PersonList: []Person{
			{0, "Charlie", 35},
			{1, "Alice", 30},
			{2, "Bob", 25},
		},
	}
}

func (s *State) Build() goui.Widget {
	return &widgets.Column{
		Widgets: []goui.Widget{

			&widgets.Button{
				Label: "--HEADER-- (ID changes on every build)",
				ID:    goui.UniqueID(),
			},

			NewPersonWidget(s.PersonList[0]),
			NewPersonWidget(s.PersonList[1]),
			NewPersonWidget(s.PersonList[2]),

			&widgets.Padding{
				Top: 20,
				Widget: &widgets.Button{
					Label: "Sort by name",
					OnClick: func(ctx *goui.Context) {
						// Update the whole Root widget to rebuild children
						gg.MustOK(s.Update(func() {
							// Sort personList by Name
							slices.SortStableFunc(s.PersonList, func(a, b Person) int {
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
						gg.MustOK(s.Update(func() {
							// Sort personList by Age
							slices.SortStableFunc(s.PersonList, func(a, b Person) int {
								return a.Age - b.Age
							})
						}))
					},
				},
			},
		},
	}
}

func Root() goui.StatefulWidget {
	return goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.State {
		return NewState(ctx)
	})
}

type PersonState struct {
	goui.StateUpdater
	goui.NopDestroyer
	person  *Person
	clicked int
}

func (s *PersonState) Build() goui.Widget {
	return &widgets.Button{
		ID:    goui.ValueID(s.person.ID),
		Label: fmt.Sprintf("%s (%d years old) - Clicked %d times", s.person.Name, s.person.Age, s.clicked),
		OnClick: func(ctx *goui.Context) {
			gg.MustOK(s.Update(func() { s.clicked++ }))
		},
	}
}

func NewPersonState(ctx *goui.StateContext, person *Person) goui.State {
	return &PersonState{
		StateUpdater: goui.NewStateUpdater(ctx),
		person:       person,
		clicked:      0,
	}
}

func NewPersonWidget(person Person) goui.StatefulWidget {
	return goui.NewStatefulWidget(
		goui.ValueID(person.ID),
		func(ctx *goui.StateContext) goui.State {
			return NewPersonState(ctx, &person)
		})

}
