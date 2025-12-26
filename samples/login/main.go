package main

import (
	"fmt"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	app.CreateWindow(goui.Window{
		OnClose: func() { app.Exit(0) },
		Title:   "goui login sample",
		Width:   400,
		Height:  300,
		Root:    rootWidget(),
	})

	app.Run()
}

const username = "admin"
const password = "password"

func rootWidget() goui.Widget {
	var userNameCtrl widgets.TextFieldController
	var passwordCtrl widgets.TextFieldController
	return &widgets.Center{
		Widget: &widgets.Column{
			CrossAxisAlignment: axes.Center,
			Widgets: []goui.Widget{
				userPass(&userNameCtrl, &passwordCtrl),
				&widgets.SizedBox{Height: 10},
				&widgets.Column{
					MainAxisSize:       axes.Min,
					CrossAxisAlignment: axes.Center,
					Widgets: []goui.Widget{
						&widgets.Button{
							Label:   "Login",
							Padding: &goui.Size{Width: 60, Height: 10},
							OnClick: func(ctx *goui.Context) {
								doLogin(ctx, &userNameCtrl, &passwordCtrl)
							},
						},
						&widgets.Padding{
							Top: 50,
							Widget: &widgets.Label{
								Text: fmt.Sprintf("Note: Use '%s' as username and '%s' as password.", username, password),
							},
						},
					},
				},
			},
		},
	}
}

func doLogin(ctx *goui.Context, userNameCtrl, passwordCtrl *widgets.TextFieldController) {
	user := gg.Must(userNameCtrl.Text())
	pass := gg.Must(passwordCtrl.Text())
	if user == username && pass == password {
		ctx.MessageBox("Login", "Logged in successfully!", goui.MessageBoxIconInfo)
	} else {
		ctx.MessageBox("Login", "Invalid username or password.", goui.MessageBoxIconError)
		userNameCtrl.SetText("")
		passwordCtrl.SetText("")
	}
}

func userPass(userNameCtrl, passwordCtrl *widgets.TextFieldController) goui.Widget {
	const rowWidth = 170
	const rowHeight = 30
	return &widgets.Column{
		MainAxisSize:       axes.Min,
		CrossAxisAlignment: axes.Center,
		Widgets: []goui.Widget{
			&widgets.SizedBox{Height: 10},
			&widgets.SizedBox{
				Width:  rowWidth,
				Height: rowHeight,
				Widget: &widgets.Row{
					CrossAxisAlignment: axes.Center,
					Widgets: []goui.Widget{
						&widgets.Expanded{
							Flex: 1,
							Widget: &widgets.Padding{
								Right: 10,
								Widget: &widgets.Label{
									Text: "Username:",
								},
							},
						},
						&widgets.SizedBox{Width: 10},
						&widgets.SizedBox{
							Width:  100,
							Height: 25,
							Widget: &widgets.TextField{
								InitialValue: username,
								Controller:   userNameCtrl,
							},
						},
					},
				},
			},
			&widgets.SizedBox{Height: 10},
			&widgets.SizedBox{
				Width:  rowWidth,
				Height: rowHeight,
				Widget: &widgets.Row{
					CrossAxisAlignment: axes.Center,
					Widgets: []goui.Widget{
						&widgets.Expanded{
							Flex: 1,
							Widget: &widgets.Label{
								Text: "Password:",
							},
						},
						&widgets.SizedBox{Width: 10},
						&widgets.SizedBox{
							Width:  100,
							Height: 25,
							Widget: &widgets.TextField{
								InitialValue: password,
								Controller:   passwordCtrl,
							},
						},
					},
				},
			},
			&widgets.SizedBox{Height: 10},
		},
	}
}
