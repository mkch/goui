package textfield

import "github.com/mkch/goui/native"

// Controller is used to control a TextField widget.
type Controller struct {
	element *textFieldElement
}

func (ctrl *Controller) setElement(elem *textFieldElement) {
	ctrl.element = elem
}

// Text returns the current text in the TextField.
func (ctrl *Controller) Text() (string, error) {
	return native.GetTextFieldText(ctrl.element.Handle)
}

// SetText sets the text in the TextField.
func (ctrl *Controller) SetText(text string) error {
	return native.SetTextFieldText(ctrl.element.Handle, text)
}
