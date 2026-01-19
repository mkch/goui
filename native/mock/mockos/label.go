package mockos

import (
	"image/color"

	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type label struct {
	abstractWindow
	multiline       bool
	textAlignment   native.TextAlignment
	backgroundColor color.NRGBA
}

func NewLabel(parent Handle, title string) (handle Handle, err error) {
	var lbl label
	initAbstractWindow(&lbl.abstractWindow, &metrics.Rect{})
	lbl.SetText(title)
	handle = newHandle()
	windows[handle] = &lbl
	if err = AddChild(parent, handle); err != nil {
		return nil, err
	}
	return handle, nil
}

func (lbl *label) SetMultiline(multiline bool) {
	lbl.multiline = multiline
}

func (lbl *label) SetTextAlignment(alignment native.TextAlignment) {
	lbl.textAlignment = alignment
}

func (lbl *label) SetBackgroundColor(clr *color.NRGBA) {
	if clr == nil {
		lbl.backgroundColor = color.NRGBA{}
		return
	}
	lbl.backgroundColor = *clr
}
