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

	statelessWidget := &testStatelessWidget{
		id:   ValueID("stateless_wrapper"),
		elem: &ElementBase{}, // This will be set to statelessElem
		menu: menuWidget,
	}

	// Create the element for the stateless widget
	statelessElem := &ElementBase{
		theWidget: statelessWidget,
	}
	statelessWidget.elem = statelessElem

	// Manually append menu element as child (simulating what the framework would do)
	element_AppendChild(statelessElem, menuElem)

	result := unwrapNativeMenu(statelessElem)

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

	statefulWidget := &testStatefulWidget{
		id:   ValueID("stateful_wrapper"),
		elem: &ElementBase{}, // This will be set to statefulElem
		menu: menuWidget,
	}

	// Create the element for the stateful widget
	statefulElem := &ElementBase{
		theWidget: statefulWidget,
	}
	statefulWidget.elem = statefulElem

	// Manually append menu element as child (simulating what the framework would do)
	element_AppendChild(statefulElem, menuElem)

	result := unwrapNativeMenu(statefulElem)

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
	statefulWidget := &testStatefulWidget{
		id:   ValueID("stateful_inner"),
		elem: &ElementBase{},
		menu: &mockNativeMenuWidget{id: ValueID("menu"), elem: menuElem},
	}
	statefulElem := &ElementBase{
		theWidget: statefulWidget,
	}
	statefulWidget.elem = statefulElem
	element_AppendChild(statefulElem, menuElem)

	// Create a StatelessWidget that contains the StatefulWidget
	statelessWidget := &testStatelessWidget{
		id:   ValueID("stateless_outer"),
		elem: &ElementBase{},
		menu: statefulWidget, // Returns the stateful widget
	}
	statelessElem := &ElementBase{
		theWidget: statelessWidget,
	}
	statelessWidget.elem = statelessElem
	element_AppendChild(statelessElem, statefulElem)

	result := unwrapNativeMenu(statelessElem)

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
	statelessWidget := &testStatelessWidget{
		id:   ValueID("stateless_inner"),
		elem: &ElementBase{},
		menu: &mockNativeMenuWidget{id: ValueID("menu"), elem: menuElem},
	}
	statelessElem := &ElementBase{
		theWidget: statelessWidget,
	}
	statelessWidget.elem = statelessElem
	element_AppendChild(statelessElem, menuElem)

	// Create a StatefulWidget that contains the StatelessWidget
	statefulWidget := &testStatefulWidget{
		id:   ValueID("stateful_outer"),
		elem: &ElementBase{},
		menu: statelessWidget, // Returns the stateless widget
	}
	statefulElem := &ElementBase{
		theWidget: statefulWidget,
	}
	statefulWidget.elem = statefulElem
	element_AppendChild(statefulElem, statelessElem)

	result := unwrapNativeMenu(statefulElem)

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
	statelessElem.theWidget = &testStatelessWidget{
		id:   ValueID("regular"),
		elem: statelessElem,
	}

	result := unwrapNativeMenu(statelessElem)

	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// TestContainsNativeMenu_DirectMenu tests containsNativeMenu when root is directly a NativeMenuElement
func TestContainsNativeMenu_DirectMenu(t *testing.T) {
	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      12345,
	}

	result := containsNativeMenu(menuElem)

	if !result {
		t.Errorf("expected true, got %v", result)
	}
}

// TestContainsNativeMenu_DeepNested tests containsNativeMenu finds menu deeply nested in tree
func TestContainsNativeMenu_DeepNested(t *testing.T) {
	// Create a container with menu inside
	containerElem := &ElementBase{}
	containerElem.theWidget = &testContainer{
		id:   ValueID("container"),
		elem: containerElem,
	}

	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      12345,
	}
	element_AppendChild(containerElem, menuElem)

	result := containsNativeMenu(containerElem)

	if !result {
		t.Errorf("expected true (menu found despite container), got %v", result)
	}
}

// TestContainsNativeMenu_NotFound tests containsNativeMenu when no menu exists
func TestContainsNativeMenu_NotFound(t *testing.T) {
	// Create a tree without any menu elements
	statelessElem := &ElementBase{}
	statelessElem.theWidget = &testStatelessWidget{
		id:   ValueID("widget"),
		elem: statelessElem,
	}

	result := containsNativeMenu(statelessElem)

	if result {
		t.Errorf("expected false, got %v", result)
	}
}

// TestContainsNativeMenu_MultipleChildren tests containsNativeMenu with multiple children (like Container)
func TestContainsNativeMenu_MultipleChildren(t *testing.T) {
	// Create a container with multiple children, one of which is a menu
	containerElem := &ElementBase{}
	containerElem.theWidget = &testContainer{
		id:   ValueID("container"),
		elem: containerElem,
	}

	// Add multiple children
	regularChild := &ElementBase{}
	element_AppendChild(containerElem, regularChild)

	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      12345,
	}
	element_AppendChild(containerElem, menuElem)

	result := containsNativeMenu(containerElem)

	if !result {
		t.Errorf("expected true (menu found among multiple children), got %v", result)
	}
}

// TestUnwrapNativeMenu_vs_ContainsNativeMenu tests the difference in behavior
// unwrapNativeMenu stops at Container, containsNativeMenu continues
func TestUnwrapNativeMenu_vs_ContainsNativeMenu(t *testing.T) {
	expectedHandle := native.Handle(12345)

	// Create a container with menu inside
	containerElem := &ElementBase{}
	containerElem.theWidget = &testContainer{
		id:   ValueID("container"),
		elem: containerElem,
	}

	menuElem := &mockNativeMenuElement{
		ElementBase: &ElementBase{},
		handle:      expectedHandle,
	}
	element_AppendChild(containerElem, menuElem)

	// unwrapNativeMenu should return nil (stopped by Container)
	unwrapResult := unwrapNativeMenu(containerElem)
	if unwrapResult != nil {
		t.Errorf("unwrapNativeMenu: expected nil, got %v", unwrapResult)
	}

	// containsNativeMenu should return true (continues searching)
	containsResult := containsNativeMenu(containerElem)
	if !containsResult {
		t.Errorf("containsNativeMenu: expected true, got %v", containsResult)
	}
}

// Test helper widgets

type testStatelessWidget struct {
	id   ID
	elem *ElementBase
	menu Widget
}

func (w *testStatelessWidget) WidgetID() ID {
	return w.id
}

func (w *testStatelessWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	return w.elem, nil
}

func (w *testStatelessWidget) Build(ctx *Context) Widget {
	return w.menu
}

func (w *testStatelessWidget) Exclusive(StatelessWidget) {}

type testStatefulWidget struct {
	id   ID
	elem *ElementBase
	menu Widget
}

func (w *testStatefulWidget) WidgetID() ID {
	return w.id
}

func (w *testStatefulWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	return w.elem, nil
}

func (w *testStatefulWidget) CreateState(ctx *StateContext) State {
	return &testState{menu: w.menu}
}

func (w *testStatefulWidget) Exclusive(StatefulWidget) {}

type testState struct {
	menu Widget
}

func (s *testState) Build() Widget {
	return s.menu
}

func (s *testState) Destroy() {}

func (s *testState) Update(updater func()) error {
	updater()
	return nil
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
	return c.elem.numChildren()
}

func (c *testContainer) Child(n int) Widget {
	return nil
}

func (c *testContainer) Exclusive(Container) {}
