package messagebox_test

import (
	"testing"

	"github.com/mkch/goui"
	"github.com/mkch/goui/gouitest"
	"github.com/mkch/goui/messagebox"
	"github.com/mkch/goui/native"
	"github.com/mkch/goui/native/mock"
)

func TestMessageBox_NoContext(t *testing.T) {
	gouitest.Run(func() {
		defer goui.Exit(0)

		os := goui.OS().(*mock.OS)
		os.Debug_SetNextMessageBoxReturn(native.MessageBoxReturnOK, nil)

		ret, err := messagebox.Show(nil, "Title", "Message", messagebox.IconInfo, messagebox.ButtonOK)
		if err != nil {
			t.Fatalf("Show returned error: %v", err)
		}
		if ret != messagebox.ReturnOK {
			t.Fatalf("Show returned %v, want %v", ret, messagebox.ReturnOK)
		}

		osReceived := os.Debug_LastMessageBoxParams()
		if osReceived.Title != "Title" {
			t.Errorf("OS received Title %q, want %q", osReceived.Title, "Title")
		}
		if osReceived.Message != "Message" {
			t.Errorf("OS received Message %q, want %q", osReceived.Message, "Message")
		}
		if osReceived.Icon != native.MessageBoxIconInfo {
			t.Errorf("OS received Icon %v, want %v", osReceived.Icon, native.MessageBoxIconInfo)
		}
		if osReceived.Button != native.MessageBoxButtonOK {
			t.Errorf("OS received Button %v, want %v", osReceived.Button, native.MessageBoxButtonOK)
		}
	}, nil)
}

func TestMessageBox(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		defer goui.Exit(0)

		os := goui.OS().(*mock.OS)
		os.Debug_SetNextMessageBoxReturn(native.MessageBoxReturnOK, nil)

		ret, err := messagebox.Show(ctx, "Title", "Message", messagebox.IconInfo, messagebox.ButtonOK)
		if err != nil {
			t.Fatalf("Show returned error: %v", err)
		}
		if ret != messagebox.ReturnOK {
			t.Fatalf("Show returned %v, want %v", ret, messagebox.ReturnOK)
		}

		osReceived := os.Debug_LastMessageBoxParams()
		if osReceived.Title != "Title" {
			t.Errorf("OS received Title %q, want %q", osReceived.Title, "Title")
		}
		if osReceived.Message != "Message" {
			t.Errorf("OS received Message %q, want %q", osReceived.Message, "Message")
		}
		if osReceived.Icon != native.MessageBoxIconInfo {
			t.Errorf("OS received Icon %v, want %v", osReceived.Icon, native.MessageBoxIconInfo)
		}
		if osReceived.Button != native.MessageBoxButtonOK {
			t.Errorf("OS received Button %v, want %v", osReceived.Button, native.MessageBoxButtonOK)
		}
	}, nil)
}
