package mockos

import "github.com/mkch/goui/metrics"

type textField struct {
	abstractWindow
}

func NewTextField(parent Handle, initialValue string, password bool) (handle Handle, err error) {
	var tf textField
	initAbstractWindow(&tf.abstractWindow, &metrics.Rect{})
	tf.SetText(initialValue)
	handle = newHandle()
	windows[handle] = &tf
	return handle, nil
}
