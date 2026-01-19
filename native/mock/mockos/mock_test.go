package mockos

import (
	"testing"

	"github.com/mkch/goui/metrics"
)

func Test_NewHandle(t *testing.T) {
	h1 := newHandle()
	if h1 == nil {
		t.Fatal("newHandle returned nil")
	}
	h2 := newHandle()
	if h2 == nil {
		t.Fatal("newHandle returned nil")
	}
	if h1 == h2 {
		t.Fatal("newHandle returned the same handle twice")
	}

	m := make(map[Handle]bool)
	for range 999 {
		m[newHandle()] = true
	}
	if len(m) != 999 {
		t.Fatal("newHandle returned duplicate handles")
	}
}

func Test_Window(t *testing.T) {
	w := initAbstractWindow(&abstractWindow{}, &metrics.Rect{})
	w.callOnDestroyListener()     // Should not panic
	w.callOnSizeChangedListener() // Should not panic
	if w.Parent() != nil {
		t.Fatal("expected nil parent")
	}
	if w.Text() != "" {
		t.Fatal("expected empty text")
	}
	if w.SetText("hello"); w.Text() != "hello" {
		t.Fatal("text not set correctly")
	}
	if len(w.Children()) != 0 {
		t.Fatal("expected no children")
	}
	if err := w.RemoveChild(nil); err == nil {
		t.Fatal("expected error when remove invalid child")
	}

	var returnValue = "return value"
	w.SetWndProc(func(msg *Msg, prev func(*Msg) any) any {
		if prev == nil {
			t.Fatal("previous WndProc is nil")
		}
		if _, ok := msg.Message.(MsgClosed); !ok {
			t.Fatal("unexpected message type")
		}
		prev(msg)
		return returnValue
	})
	result := w.CallWndProc(&Msg{Message: MsgClosed{}})
	if result != returnValue {
		t.Fatal("WndProc did not return expected value")
	}
}
