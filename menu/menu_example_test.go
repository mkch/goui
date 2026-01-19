package menu_test

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/menu"
)

func ExampleMenu() {
	goui.Run(func() {
		goui.CreateWindow(&goui.Window{
			Title:  "Menu Example",
			Width:  400,
			Height: 300,
			Menu: &menu.Menu{
				Items: []goui.MenuItem{&menu.Item{
					Title: "File",
					Submenu: menu.Items{
						&menu.Item{
							Title:    "New",
							OnSelect: func(ctx *goui.Context) { /* handle New action */ },
						},
						&menu.Separator{},
						&menu.Item{
							Title:    "Exit",
							OnSelect: func(ctx *goui.Context) { goui.Exit(0) },
						},
					},
				}},
			},
		})
	}, nil)

}
