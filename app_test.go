package goui

import (
	"testing"

	"github.com/mkch/goui/native"
)

// mockNativeMenuWidget wraps a NativeMenuElement to implement Widget
type mockNativeMenuWidget struct {
	id   ID
	elem *mockNativeMenuElement
}

func (w *mockNativeMenuWidget) WidgetID() ID {
	return w.id
}

func (w *mockNativeMenuWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	return w.elem, nil
}

// mockNativeMenuElement wraps ElementBase to add NativeMenuElement behavior for testing
type mockNativeMenuElement struct {
	*ElementBase
	handle native.Handle
}

func (m *mockNativeMenuElement) NativeMenu() native.Handle {
	return m.handle
}

// TestUnwrapNativeMenu_DirectMenu tests unwrapNativeMenu when the root element is directly a NativeMenuElement
func TestUnwrapNativeMenu_DirectMenu(t *testing.T) {
	expectedHandle := native.Handle(12345)
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}

	result := unwrapNativeMenu(menuElem)

	if result != expectedHandle {
		t.Errorf("expected handle %v, got %v", expectedHandle, result)
	}
}

// TestUnwrapNativeMenu_WrappedInStatelessWidget tests unwrapNativeMenu finding menu wrapped in StatelessWidget
func TestUnwrapNativeMenu_WrappedInStatelessWidget(t *testing.T) {
	expectedHandle := native.Handle(12345)

	// Create the menu element
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}

	// Create a StatelessWidget that builds to the menu
	menuWidget := &mockNativeMenuWidget{
		id:   ValueID("menu"),
		elem: menuElem,
	}

	statelessWidget := &StatelessWidget{
		ID: ValueID("stateless_wrapper"),
		Builder: func(ctx *Context) Widget {
			return menuWidget
		},
	}

	elem, err := BuildElementTree(newMockContext(nil), statelessWidget)
	if err != nil {
		t.Fatalf("failed to build element tree: %v", err)
	}

	result := unwrapNativeMenu(elem)

	if result != expectedHandle {
		t.Errorf("expected handle %v, got %v", expectedHandle, result)
	}
}

// TestUnwrapNativeMenu_WrappedInStatefulWidget tests unwrapNativeMenu finding menu wrapped in StatefulWidget
func TestUnwrapNativeMenu_WrappedInStatefulWidget(t *testing.T) {
	expectedHandle := native.Handle(12345)

	// Create the menu element
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}

	// Create a StatefulWidget that builds to the menu
	menuWidget := &mockNativeMenuWidget{
		id:   ValueID("menu"),
		elem: menuElem,
	}

	statefulWidget := &StatefulWidget{
		ID: ValueID("stateful_wrapper"),
		StateCreator: func(ctx *StateContext) State {
			return NewState(ctx, func() Widget {
				return menuWidget
			}, nil)
		},
	}

	elem, err := BuildElementTree(newMockContext(nil), statefulWidget)
	if err != nil {
		t.Fatalf("failed to build element tree: %v", err)
	}

	result := unwrapNativeMenu(elem)

	if result != expectedHandle {
		t.Errorf("expected handle %v, got %v", expectedHandle, result)
	}
}

// TestUnwrapNativeMenu_StatefulInStateless tests StatefulWidget nested in StatelessWidget with menu
func TestUnwrapNativeMenu_StatefulInStateless(t *testing.T) {
	expectedHandle := native.Handle(12345)

	// Create the menu element
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}

	// Create a StatefulWidget that has the menu
	statefulWidget := &StatefulWidget{
		ID: ValueID("stateful_inner"),
		StateCreator: func(ctx *StateContext) State {
			return NewState(ctx, func() Widget {
				return &mockNativeMenuWidget{id: ValueID("menu"), elem: menuElem}
			}, nil)
		},
	}

	// Create a StatelessWidget that contains the StatefulWidget
	statelessWidget := &StatelessWidget{
		ID: ValueID("stateless_outer"),
		Builder: func(ctx *Context) Widget {
			return statefulWidget
		},
	}

	elem, err := BuildElementTree(newMockContext(nil), statelessWidget)
	if err != nil {
		t.Fatalf("failed to build element tree: %v", err)
	}

	result := unwrapNativeMenu(elem)

	if result != expectedHandle {
		t.Errorf("expected handle %v, got %v", expectedHandle, result)
	}
}

// TestUnwrapNativeMenu_StatelessInStateful tests StatelessWidget nested in StatefulWidget with menu
func TestUnwrapNativeMenu_StatelessInStateful(t *testing.T) {
	expectedHandle := native.Handle(12345)

	// Create the menu element
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}

	// Create a StatelessWidget that has the menu
	statelessWidget := &StatelessWidget{
		ID: ValueID("stateless_inner"),
		Builder: func(ctx *Context) Widget {
			return &mockNativeMenuWidget{id: ValueID("menu"), elem: menuElem}
		},
	}

	// Create a StatefulWidget that contains the StatelessWidget
	statefulWidget := &StatefulWidget{
		ID: ValueID("stateful_outer"),
		StateCreator: func(ctx *StateContext) State {
			return NewState(ctx, func() Widget {
				return statelessWidget
			}, nil)
		},
	}

	elem, err := BuildElementTree(newMockContext(nil), statefulWidget)
	if err != nil {
		t.Fatalf("failed to build element tree: %v", err)
	}

	result := unwrapNativeMenu(elem)

	if result != expectedHandle {
		t.Errorf("expected handle %v, got %v", expectedHandle, result)
	}
}

// TestUnwrapNativeMenu_BlockedByContainer tests that unwrapNativeMenu stops at Container
func TestUnwrapNativeMenu_BlockedByContainer(t *testing.T) {
	expectedHandle := native.Handle(12345)

	// Create a container (which is not StatelessWidget or StatefulWidget)
	containerElem := &ElementBase{}
	containerElem.theWidget = &testContainer{
		id:   ValueID("container"),
		elem: containerElem,
	}

	// Create menu element as child
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}
	element_AppendChild(containerElem, menuElem)

	result := unwrapNativeMenu(containerElem)

	if result != nil {
		t.Errorf("expected nil (blocked by container), got %v", result)
	}
}

// TestUnwrapNativeMenu_NotFound tests unwrapNativeMenu when no menu is found
func TestUnwrapNativeMenu_NotFound(t *testing.T) {
	// Create a stateless widget without a menu
	statelessElem := &ElementBase{}
	statelessElem.theWidget = &StatelessWidget{
		ID: ValueID("regular"),
		Builder: func(ctx *Context) Widget {
			return nil
		},
	}

	result := unwrapNativeMenu(statelessElem)

	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

type testContainer struct {
	id   ID
	elem *ElementBase
}

func (c *testContainer) WidgetID() ID {
	return c.id
}

func (c *testContainer) CreateElement(ctx *Context, parent Element) (Element, error) {
	return c.elem, nil
}

func (c *testContainer) NumChildren() int {
	return c.elem.NumChildren()
}

func (c *testContainer) Child(n int) Widget {
	return nil
}

func (c *testContainer) Exclusive(Container) {}
