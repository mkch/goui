package textfield

import (
	"github.com/mkch/goui"
)

// Controller is used to control a TextField widget.
type Controller struct {
	ctx     *goui.Context
	element *textFieldElement
}

func (ctrl *Controller) setElement(ctx *goui.Context, elem *textFieldElement) {
	ctrl.ctx = ctx
	ctrl.element = elem
}

// Text returns the current text in the TextField.
func (ctrl *Controller) Text() (string, error) {
	return goui.OS().TextField_Text(ctrl.element.Handle)
}

// SetText sets the text in the TextField.
func (ctrl *Controller) SetText(text string) error {
	return goui.OS().TextField_SetText(ctrl.element.Handle, text)
}
