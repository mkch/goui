package main

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp()

func main() {
	app.CreateWindow(goui.Window{
		DebugLayout: true,
		OnClose:     func() { app.Exit(0) },
		Title:       "goui login sample",
		Width:       400,
		Height:      300,
		Root:        rootWidget(),
	})

	app.Run()
}

func rootWidget() goui.Widget {
	var userNameCtrl widgets.TextFieldController
	var passwordCtrl widgets.TextFieldController
	return &widgets.Column{
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
						OnClick: func() {
							doLogin(&userNameCtrl, &passwordCtrl)
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
	}
}

func doLogin(userNameCtrl, passwordCtrl *widgets.TextFieldController) {
	user := gg.Must(userNameCtrl.Text())
	pass := gg.Must(passwordCtrl.Text())
	if user == "admin" && pass == "password" {
		app.MessageBox("Login", "Logged in successfully!", goui.MessageBoxIconInfo)
	} else {
		app.MessageBox("Login", "Invalid username or password.", goui.MessageBoxIconError)
		userNameCtrl.SetText("")
		passwordCtrl.SetText("")
	}
}

func userPass(userNameCtrl, passwordCtrl *widgets.TextFieldController) goui.Widget {
	return &widgets.Column{
		MainAxisSize: axes.Min,
		Widgets: []goui.Widget{
			&widgets.SizedBox{Height: 10},
			&widgets.Row{
				MainAxisSize:       axes.Min,
				CrossAxisAlignment: axes.Center,
				Widgets: []goui.Widget{
					&widgets.Padding{
						Right: 10,
						Widget: &widgets.Label{
							Text: "Username:",
						},
					},
					&widgets.Padding{
						Right: 10,
						Widget: &widgets.SizedBox{
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
			&widgets.Row{
				MainAxisSize: axes.Min,
				Widgets: []goui.Widget{
					&widgets.Padding{
						Right: 10,
						Widget: &widgets.Label{
							Text: "Password:",
						},
					},
					&widgets.Padding{
						Right: 10,
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
			&widgets.SizedBox{Height: 10},
		},
	}
}
