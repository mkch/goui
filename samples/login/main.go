package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/messagebox"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	app.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "goui login sample",
		Width:     800,
		Height:    600,
		Root:      rootWidget(),
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
				&widgets.SizedBox{Height: 20},
				&widgets.Column{
					MainAxisSize:       axes.Min,
					CrossAxisAlignment: axes.Center,
					Widgets: []goui.Widget{
						&widgets.Button{
							Label:   "Login",
							Padding: &goui.Size{Width: 120, Height: 20},
							OnClick: func(ctx *goui.Context) {
								doLogin(ctx, &userNameCtrl, &passwordCtrl)
							},
						},
						&widgets.Padding{
							Top: 100,
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
		messagebox.Show(ctx, "Login", "Logged in successfully!", messagebox.IconInfo, messagebox.ButtonOK)
	} else {
		messagebox.Show(ctx, "Login", "Invalid username or password.", messagebox.IconError, messagebox.ButtonOK)
		userNameCtrl.SetText("")
		passwordCtrl.SetText("")
	}
}

func userPass(userNameCtrl, passwordCtrl *widgets.TextFieldController) goui.Widget {
	const rowWidth = 300
	const rowHeight = 60
	return &widgets.Column{
		MainAxisSize:       axes.Min,
		CrossAxisAlignment: axes.Center,
		Widgets: []goui.Widget{
			&widgets.SizedBox{Height: 20},
			&widgets.SizedBox{
				Width:  rowWidth,
				Height: rowHeight,
				Widget: &widgets.Row{
					CrossAxisAlignment: axes.Center,
					Widgets: []goui.Widget{
						&widgets.Expanded{
							Flex: 1,
							Widget: &widgets.Label{
								Text: "Username:",
							},
						},
						&widgets.SizedBox{Width: 20},
						&widgets.SizedBox{
							Width:  150,
							Height: 40,
							Widget: &widgets.TextField{
								InitialValue: username,
								Controller:   userNameCtrl,
							},
						},
					},
				},
			},
			&widgets.SizedBox{Height: 20},
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
						&widgets.SizedBox{Width: 20},
						&widgets.SizedBox{
							Width:  150,
							Height: 40,
							Widget: &widgets.TextField{
								Obscure:      true,
								InitialValue: password,
								Controller:   passwordCtrl,
							},
						},
					},
				},
			},
			&widgets.SizedBox{Height: 20},
		},
	}
}
