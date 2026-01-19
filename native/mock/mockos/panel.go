package mockos

import (
	"image/color"

	"github.com/mkch/goui/metrics"
)

type panel struct {
	abstractWindow
	backgroundColor color.NRGBA
}

func NewPanel(parent Handle) (handle Handle, err error) {
	var p panel
	initAbstractWindow(&p.abstractWindow, &metrics.Rect{})
	handle = newHandle()
	windows[handle] = &p
	if err = AddChild(parent, handle); err != nil {
		return nil, err
	}
	return handle, nil
}

func (p *panel) SetBackgroundColor(clr *color.NRGBA) {
	if clr == nil {
		p.backgroundColor = color.NRGBA{}
		return
	}
	p.backgroundColor = *clr
}
