package goui

import (
	"errors"
	"testing"
)

type mockWidget struct {
	ID          ID
	createError error
	element     Element
}

func (w *mockWidget) WidgetID() ID {
	return w.ID
}

func (w *mockWidget) CreateElement(ctx *Context) (Element, error) {
	if w.createError != nil {
		return nil, w.createError
	}
	return w.element, nil
}

func TestBuildElementTree_CreateElementError(t *testing.T) {
	ctx := &Context{}
	expectedErr := errors.New("create element error")
	widget := &mockWidget{
		createError: expectedErr,
	}

	elem, layouter, err := buildElementTree(ctx, widget, nil)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if elem != nil {
		t.Errorf("expected nil element, got %v", elem)
	}
	if layouter != nil {
		t.Errorf("expected nil layouter, got %v", layouter)
	}
}

func TestBuildElementTree_SimpleWidget(t *testing.T) {
	ctx := &Context{}
	widget := &mockWidget{ID: ValueID("test"), element: &element{}}

	elem, layouter, err := buildElementTree(ctx, widget, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.widget() != widget {
		t.Errorf("element widget not set correctly")
	}
	if elem.numChildren() != 0 {
		t.Errorf("expected 0 children, got %d", elem.numChildren())
	}
	if layouter != nil {
		t.Errorf("expected nil layouter for simple widget, got %v", layouter)
	}
}

type mockLayouter struct {
	LayouterBase
}

func (l *mockLayouter) Layout(ctx *Context, constraints Constraints) Size {
	return Size{Width: 100, Height: 100}
}

func (l *mockLayouter) Apply(x, y int) error {
	return nil
}

type mockElement struct {
	element
	layouter Layouter
}

func (e *mockElement) Layouter() Layouter {
	return e.layouter
}

func TestBuildElementTree_WidgetWithLayouter(t *testing.T) {
	ctx := &Context{}
	mockLayouter := &mockLayouter{}
	mockElement := &mockElement{
		layouter: mockLayouter,
	}
	mockWidget := &mockWidget{ID: ValueID("test"), element: mockElement}

	resultElem, layouter, err := buildElementTree(ctx, mockWidget, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resultElem != mockElement {
		t.Errorf("unexpected element")
	}
	if widget := resultElem.widget(); widget != mockWidget {
		t.Errorf("element widget not set correctly")
	}
	if layouter != mockLayouter {
		t.Errorf("expected layouter to be returned")
	}
	if mockLayouter.element() != mockElement {
		t.Errorf("layouter element not set correctly")
	}
}

func TestBuildElementTree_StatefulWidget(t *testing.T) {
	ctx := &Context{}
	mockLayouter := &mockLayouter{}
	mockElement := &mockElement{
		layouter: mockLayouter,
	}
	childWidget := &mockWidget{ID: ValueID("child"), element: mockElement}

	widget := NewStatefulWidget(ValueID("stateful"), func(ctx *Context) *WidgetState {
		return &WidgetState{Build: func() Widget { return childWidget }}
	})

	elem, layouter, err := buildElementTree(ctx, widget, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.widget().WidgetID() != widget.WidgetID() {
		t.Errorf("element widget not set correctly")
	}
	if elem.numChildren() != 1 {
		t.Errorf("expected 1 child, got %d", elem.numChildren())
	}
	if elem.child(0).widget().WidgetID() != childWidget.WidgetID() {
		t.Errorf("child widget not set correctly")
	}
	if layouter != mockLayouter {
		t.Errorf("wrong layouter returned")
	}
}

func TestBuildElementTree_StatelessWidget(t *testing.T) {
	ctx := &Context{}
	mockLayouter := &mockLayouter{}
	mockElement := &mockElement{
		layouter: mockLayouter,
	}
	childWidget := &mockWidget{ID: ValueID("child"), element: mockElement}

	widget := NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget {
		return childWidget
	})

	elem, layouter, err := buildElementTree(ctx, widget, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.widget().WidgetID() != widget.WidgetID() {
		t.Errorf("element widget not set correctly")
	}
	if elem.numChildren() != 1 {
		t.Errorf("expected 1 child, got %d", elem.numChildren())
	}
	if elem.child(0).widget().WidgetID() != childWidget.WidgetID() {
		t.Errorf("child widget not set correctly")
	}
	if layouter != mockLayouter {
		t.Errorf("wrong layouter returned")
	}
}

type mockContainer struct {
	ID       ID
	Children []Widget
}

func (c *mockContainer) WidgetID() ID {
	return c.ID
}

func (c *mockContainer) CreateElement(ctx *Context) (Element, error) {
	return &mockElement{layouter: &mockLayouter{}}, nil
}

func (c *mockContainer) NumChildren() int {
	return len(c.Children)
}

func (c *mockContainer) Child(n int) Widget {
	return c.Children[n]
}

func TestBuildElementTree_Container(t *testing.T) {
	ctx := &Context{}
	layouter1 := &mockLayouter{}
	child1 := &mockWidget{ID: ValueID("child1"), element: &mockElement{layouter: layouter1}}
	layouter2 := &mockLayouter{}
	child2 := &mockWidget{ID: ValueID("child2"), element: &mockElement{layouter: layouter2}}

	container := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{child1, NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget {
			return child2
		})},
	}

	elem, layouter, err := buildElementTree(ctx, container, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.numChildren() != 2 {
		t.Errorf("expected 2 children, got %d", elem.numChildren())
	}
	if elem.child(0).widget() != child1 {
		t.Errorf("first child widget not set correctly")
	}
	if elem.child(1).widget().WidgetID() != ValueID("stateless") {
		t.Errorf("second child widget not set correctly")
	}

	if layouter == nil {
		t.Errorf("expected non-nil layouter for container widget")
	}
	if layouter.numChildren() != 2 {
		t.Errorf("layouter should have 2 children, got %d", layouter.numChildren())
	}
	if layouter.child(0) != layouter1 {
		t.Errorf("first child layouter not set correctly")
	}
	if layouter.child(1) != layouter2 {
		t.Errorf("second child layouter not set correctly")
	}
}

func TestBuildElementTree_ChildNoLayouter(t *testing.T) {
	ctx := &Context{}
	childWidget := &mockWidget{ID: ValueID("child"), element: &element{}}
	container := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget {
			return childWidget
		})},
	}
	elem, layouter, err := buildElementTree(ctx, container, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem.widget() != container {
		t.Errorf("container widget not set correctly")
	}
	if elem.numChildren() != 1 {
		t.Errorf("expected 1 child, got %d", elem.numChildren())
	}
	if layouter == nil {
		t.Errorf("expected non-nil layouter for container widget")
	}
	if layouter.numChildren() != 0 {
		t.Errorf("layouter should have 0 children, got %d", layouter.numChildren())
	}
}

func TestUpdateElementTree_Update(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &element{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &mockElement{layouter: &mockLayouter{}}}

	container1 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child2},
	}

	ctx := &Context{}
	elem, layouter, err := buildElementTree(ctx, container1, nil)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	container2 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			NewStatelessWidget(nil, func(ctx *Context) Widget { return child1 }),
			child2},
	}

	newElem, newLayouter, err := updateElementTree(ctx, elem, container2)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}

	// The root element and layouter should be the same.
	if newElem != elem {
		t.Fatalf("expected root element to be the same")
	}
	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}
	// The first child element should be replaced.
	if newElem.numChildren() != 2 {
		t.Fatalf("expected 2 children, got %d", newElem.numChildren())
	}
	if childWidget1, ok := newElem.child(0).widget().(StatelessWidget); !ok {
		t.Fatal("expected first child to be a StatelessWidget")
	} else if childWidget1.Build(ctx) != child1 {
		t.Fatal("first child widget not updated correctly")
	}
	// The second child element should be the same.
	if child2 := newElem.child(1); child2 != child2 {
		t.Fatalf("second child element not updated correctly")
	}

	// The entire layouter tree should be the same.
	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}
	if newLayouter.numChildren() != layouter.numChildren() || newLayouter.numChildren() != 1 {
		t.Fatalf("expected layouter to have same number of children")
	}
	if newLayouter.child(0) != layouter.child(0) {
		t.Fatalf("first child layouter not the same")
	}

	container3 := &mockContainer{
		Children: []Widget{
			NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget { return child1 }),
			child2},
	}

	newElem2, newLayouter2, err := updateElementTree(ctx, newElem, container3)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}
	// The root element should be recreated.
	if newElem2 == newElem {
		t.Fatalf("expected root element to be recreated")
	}
	if newLayouter2 == newLayouter {
		t.Fatalf("expected root layouter to be recreated")
	}
	if newElem2.widget() != container3 {
		t.Fatalf("new root element widget not set correctly")
	}
	if newElem2.numChildren() != 2 {
		t.Fatalf("expected 2 children, got %d", newElem2.numChildren())
	}
	if newElem2.child(0).widget().WidgetID() != ValueID("stateless") {
		t.Fatal("first child widget not updated correctly")
	}
	if newElem2.child(1).widget() != child2 {
		t.Fatalf("second child element not updated correctly")
	}

	if newLayouter2.numChildren() != 1 {
		t.Fatalf("expected layouter to have 1 child")
	}
	if newLayouter2.child(0) != layouter.child(0) {
		t.Fatalf("first child layouter not the same")
	}
}

func TestUpdateElementTree_UpdateID(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &element{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &mockElement{layouter: &mockLayouter{}}}
	child3 := &mockWidget{ID: ValueID("child3"), element: &element{}}

	container1 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			NewStatelessWidget(nil, func(ctx *Context) Widget { return child1 }),
			NewStatefulWidget(nil, func(ctx *Context) *WidgetState {
				return &WidgetState{
					Build: func() Widget { return child2 }}
			}),
			NewStatefulWidget(nil, func(ctx *Context) *WidgetState {
				return &WidgetState{
					Build: func() Widget { return child3 }}
			}),
		},
	}

	ctx := &Context{}
	elem, layouter, err := buildElementTree(ctx, container1, nil)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	if elem.numChildren() != 3 {
		t.Fatalf("expected 3 children, got %d", elem.numChildren())
	}

	child4 := &mockWidget{ID: ValueID("child4"), element: &element{}}
	child5 := &mockWidget{ID: ValueID("child5"), element: &mockElement{layouter: &mockLayouter{}}}
	container2 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			NewStatelessWidget(nil, func(ctx *Context) Widget { return child4 }), // Build() method of StatelessWidget is always called.
			NewStatefulWidget(ValueID("2"), func(ctx *Context) *WidgetState { // ID changed, so CreateState() is called.
				return &WidgetState{
					Build: func() Widget { return child5 }}
			}),
			NewStatefulWidget(nil, func(ctx *Context) *WidgetState { // Neither ID or type changed, so CreateState() will not called.
				return &WidgetState{
					Build: func() Widget { panic("should not be called") }}
			}),
		},
	}

	newElem, newLayouter, err := updateElementTree(ctx, elem, container2)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}

	// The root element and layouter should be the same.
	if newElem != elem {
		t.Fatalf("expected root element to be the same")
	}
	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}

	// The child elements should be replaced.
	if newElem.numChildren() != elem.numChildren() {
		t.Fatalf("expected same number of children, got %d and %d", newElem.numChildren(), elem.numChildren())
	}
	if id := newElem.child(0).child(0).widget().WidgetID(); id != ValueID("child4") {
		t.Fatalf("expected first child element to be replaced, got %v", id)
	}
	if id := newElem.child(1).child(0).widget().WidgetID(); id != ValueID("child5") {
		t.Fatalf("expected second child element to be replaced, got %v", id)
	}
	if id := newElem.child(2).child(0).widget().WidgetID(); id != ValueID("child3") {
		t.Fatalf("expected third child element to be the same, got %v", id)
	}
}

func TestUpdateElementTree_Append(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &element{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &mockElement{layouter: &mockLayouter{}}}
	child3 := &mockWidget{ID: ValueID("child3"), element: &mockElement{layouter: &mockLayouter{}}}

	container1 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child2},
	}

	ctx := &Context{}
	elem, layouter, err := buildElementTree(ctx, container1, nil)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	container2 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			child1, child2, child3},
	}

	newElem, newLayouter, err := updateElementTree(ctx, elem, container2)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}

	if newElem != elem {
		t.Fatalf("expected root element to be the same")
	}
	if newElem.numChildren() != 3 {
		t.Fatalf("expected 3 children, got %d", newElem.numChildren())
	}
	if newElem.child(0).widget() != child1 || newElem.child(1).widget() != child2 || newElem.child(2).widget() != child3 {
		t.Fatalf("child elements not updated correctly")
	}

	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}
	if newLayouter.numChildren() != 2 {
		t.Fatalf("expected layouter to have 2 children")
	}
	if newLayouter.child(0).element().widget() != child2 || newLayouter.child(1).element().widget() != child3 {
		t.Fatalf("child layouters not updated correctly")
	}
}

func TestUpdateElementTree_Remove(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &element{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &mockElement{layouter: &mockLayouter{}}}
	child3 := &mockWidget{ID: ValueID("child3"), element: &mockElement{layouter: &mockLayouter{}}}

	container1 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child2, child3},
	}

	ctx := &Context{}
	elem, layouter, err := buildElementTree(ctx, container1, nil)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	container2 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child3},
	}

	newElem, newLayouter, err := updateElementTree(ctx, elem, container2)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}

	if newElem != elem {
		t.Fatalf("expected root element to be the same")
	}
	if newElem.numChildren() != 2 {
		t.Fatalf("expected 2 children, got %d", newElem.numChildren())
	}
	if newElem.child(0).widget() != child1 || newElem.child(1).widget() != child3 {
		t.Fatalf("child elements not updated correctly")
	}

	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}
	if newLayouter.numChildren() != 1 {
		t.Fatalf("expected layouter to have 1 child")
	}
	if newLayouter.child(0).element().widget() != child3 {
		t.Fatalf("child layouters not updated correctly")
	}
}
