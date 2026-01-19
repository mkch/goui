package mockos

import "github.com/mkch/goui/metrics"

type button struct {
	abstractWindow
	onClick func()
}

func NewButton(parent Handle, title string) (handle Handle, err error) {
	var btn button
	initAbstractWindow(&btn.abstractWindow, &metrics.Rect{})
	btn.SetText(title)
	btn.SetWndProc(func(msg *Msg, prev func(*Msg) any) any {
		switch msg.Message.(type) {
		case MsgMouseLeftDown:
			btn.callOnClickListener()
		}
		return prev(msg)
	})

	handle = newHandle()
	windows[handle] = &btn
	if err = AddChild(parent, handle); err != nil {
		return nil, err
	}
	return handle, nil
}

func (btn *button) callOnClickListener() {
	if btn.onClick != nil {
		btn.onClick()
	}
}

func (btn *button) SetOnClickListener(onClick func()) {
	btn.onClick = onClick
}

func (btn *button) MinimumSize() (size metrics.Size) {
	label := btn.Text()
	if len(label) == 0 {
		return
	}
	size.Height = 30
	size.Width = metrics.DP(len(label))*15 + 20
	return
}
