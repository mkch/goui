package main

import (
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
								Text: "Note: Use 'admin' as username and 'password' as password.",
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
	if user == "admin" && pass == "password" {
		ctx.MessageBox("Login", "Logged in successfully!", goui.MessageBoxIconInfo)
	} else {
		ctx.MessageBox("Login", "Invalid username or password.", goui.MessageBoxIconError)
		userNameCtrl.SetText("")
		passwordCtrl.SetText("")
	}
}

func userPass(userNameCtrl, passwordCtrl *widgets.TextFieldController) goui.Widget {
	const ROW_WIDTH = 170
	const ROW_HEIGHT = 30
	return &widgets.Column{
		MainAxisSize:       axes.Min,
		CrossAxisAlignment: axes.Center,
		Widgets: []goui.Widget{
			&widgets.SizedBox{Height: 10},
			&widgets.SizedBox{
				Width:  ROW_WIDTH,
				Height: ROW_HEIGHT,
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
								Controller: userNameCtrl,
							},
						},
					},
				},
			},
			&widgets.SizedBox{Height: 10},
			&widgets.SizedBox{
				Width:  ROW_WIDTH,
				Height: ROW_HEIGHT,
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
						&widgets.Padding{
							Widget: &widgets.SizedBox{
								Width:  100,
								Height: 25,
								Widget: &widgets.TextField{
									Controller: passwordCtrl,
								},
							},
						},
					},
				},
			},
			&widgets.SizedBox{Height: 10},
		},
	}
}
