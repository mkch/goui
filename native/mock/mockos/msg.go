package mockos

import "github.com/mkch/goui/metrics"

// Message is a window message.
type Message interface {
	Message()
}

// InputMessage is an input-related window message.
type InputMessage interface {
	InputMessage()
}

type MsgPostFunc func()

func (MsgPostFunc) Message() {}

type MsgQuit int

func (MsgQuit) Message() {}

type MsgDestroying struct{}

func (MsgDestroying) Message() {}

type MsgDestroyed struct{}

func (MsgDestroyed) Message() {}

type MsgClosing struct{}

func (MsgClosing) Message() {}

type MsgClosed struct{}

func (MsgClosed) Message() {}

type MsgMouseLeftDown struct{ X, Y metrics.DP }

func (MsgMouseLeftDown) Message()      {}
func (MsgMouseLeftDown) InputMessage() {}

type MsgMouseLeftUp struct{ X, Y metrics.DP }

func (MsgMouseLeftUp) Message()      {}
func (MsgMouseLeftUp) InputMessage() {}

type MsgMouseRightDown struct{ X, Y metrics.DP }

func (MsgMouseRightDown) Message()      {}
func (MsgMouseRightDown) InputMessage() {}

type MsgMouseRightUp struct{ X, Y metrics.DP }

func (MsgMouseRightUp) Message()      {}
func (MsgMouseRightUp) InputMessage() {}

type MsgMouseMiddleDown struct{ X, Y metrics.DP }

func (MsgMouseMiddleDown) Message()      {}
func (MsgMouseMiddleDown) InputMessage() {}

type MsgMouseMiddleUp struct{ X, Y metrics.DP }

func (MsgMouseMiddleUp) Message()      {}
func (MsgMouseMiddleUp) InputMessage() {}

type MsgPaint metrics.Rect

func (MsgPaint) Message() {}

type MsgSizedChanged struct{}

func (MsgSizedChanged) Message() {}
