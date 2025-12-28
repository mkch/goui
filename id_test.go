package goui_test

import (
	"testing"

	"github.com/mkch/goui"
)

func TestValueID(t *testing.T) {
	id1, id2 := goui.ValueID(1), goui.ValueID(1)
	if id1 != id2 {
		t.Errorf("Expected ValueID(1) to be equal to ValueID(1)") // same values
	}

	id1, id2 = goui.ValueID("a"), goui.ValueID("b")
	if id1 == id2 {
		t.Errorf("Expected ValueID(\"a\") to be not equal to ValueID(\"b\")") // different values
	}

	id1, id2 = goui.ValueID(struct{ int }{1}), goui.ValueID(struct{ int }{1})
	if id1 != id2 {
		t.Errorf("Expected ValueID(struct{int}{1}) to be equal to ValueID(struct{int}{1})")
	}

	type S struct{ int }
	if id1 == goui.ValueID(S{1}) {
		t.Errorf("Expected ValueID(struct{int}{1}) to be not equal to ValueID(S{1})") // different types
	}

	id1, id2 = goui.ValueID(new(int)), goui.ValueID(new(int))
	if id1 == id2 {
		t.Errorf("Expected ValueID(new(int)) to be not equal to ValueID(new(int))") // different pointers
	}
}

func TestUniqueID(t *testing.T) {
	id1, id2 := goui.UniqueID(), goui.UniqueID()
	if id1 == id2 {
		t.Errorf("Expected UniqueID() to be not equal to UniqueID()")
	}

	for range 9999 {
		if id1 == goui.UniqueID() {
			t.Errorf("Expected UniqueID() to be unique")
		}
	}
}
