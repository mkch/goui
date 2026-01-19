package goui

import (
	"github.com/mkch/goui/native"
	"github.com/mkch/goui/native/windows"
)

func newOS() native.OS {
	return windows.NewOS()
}
