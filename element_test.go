package goui

import (
	"errors"
	"slices"
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
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	expectedErr := errors.New("create element error")
	widget := &mockWidget{
		createError: expectedErr,
	}

	elem, layouter, err := buildElementTree(ctx, widget)

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
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	widget := &mockWidget{ID: ValueID("test"), element: &ElementBase{}}

	elem, layouter, err := buildElementTree(ctx, widget)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.Widget() != widget {
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

func (l *mockLayouter) Layout(ctx *Context, constraints Constraints) (Size, error) {
	return Size{Width: 100, Height: 100}, nil
}

func (l *mockLayouter) PositionAt(x, y int) error {
	return nil
}

func TestBuildElementTree_WidgetWithLayouter(t *testing.T) {
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	mockLayouter := &mockLayouter{}
	mockElement := &ElementBase{
		ElementLayouter: mockLayouter,
	}
	mockWidget := &mockWidget{ID: ValueID("test"), element: mockElement}

	resultElem, layouter, err := buildElementTree(ctx, mockWidget)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resultElem != mockElement {
		t.Errorf("unexpected element")
	}
	if widget := resultElem.Widget(); widget != mockWidget {
		t.Errorf("element widget not set correctly")
	}
	if layouter != mockLayouter {
		t.Errorf("expected layouter to be returned")
	}
	if mockLayouter.Element() != mockElement {
		t.Errorf("layouter element not set correctly")
	}
}

func TestBuildElementTree_StatefulWidget(t *testing.T) {
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	mockLayouter := &mockLayouter{}
	mockElement := &ElementBase{
		ElementLayouter: mockLayouter,
	}
	childWidget := &mockWidget{ID: ValueID("child"), element: mockElement}

	widget := NewStatefulWidget(ValueID("stateful"), func(ctx *Context, updateState UpdateStateFunc) *WidgetState {
		return &WidgetState{Build: func() Widget { return childWidget }}
	})

	elem, layouter, err := buildElementTree(ctx, widget)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.Widget().WidgetID() != widget.WidgetID() {
		t.Errorf("element widget not set correctly")
	}
	if elem.numChildren() != 1 {
		t.Errorf("expected 1 child, got %d", elem.numChildren())
	}
	if elem.child(0).Widget().WidgetID() != childWidget.WidgetID() {
		t.Errorf("child widget not set correctly")
	}
	if layouter != mockLayouter {
		t.Errorf("wrong layouter returned")
	}
}

func TestBuildElementTree_StatelessWidget(t *testing.T) {
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	mockLayouter := &mockLayouter{}
	mockElement := &ElementBase{
		ElementLayouter: mockLayouter,
	}
	childWidget := &mockWidget{ID: ValueID("child"), element: mockElement}

	widget := NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget {
		return childWidget
	})

	elem, layouter, err := buildElementTree(ctx, widget)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.Widget().WidgetID() != widget.WidgetID() {
		t.Errorf("element widget not set correctly")
	}
	if elem.numChildren() != 1 {
		t.Errorf("expected 1 child, got %d", elem.numChildren())
	}
	if elem.child(0).Widget().WidgetID() != childWidget.WidgetID() {
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
	return &ElementBase{ElementLayouter: &mockLayouter{}}, nil
}

func (c *mockContainer) NumChildren() int {
	return len(c.Children)
}

func (c *mockContainer) Child(n int) Widget {
	return c.Children[n]
}

func (c *mockContainer) Exclusive(Container) { /*Nop*/ }

func TestBuildElementTree_Container(t *testing.T) {
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	layouter1 := &mockLayouter{}
	child1 := &mockWidget{ID: ValueID("child1"), element: &ElementBase{ElementLayouter: layouter1}}
	layouter2 := &mockLayouter{}
	child2 := &mockWidget{ID: ValueID("child2"), element: &ElementBase{ElementLayouter: layouter2}}

	container := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{child1, NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget {
			return child2
		})},
	}

	elem, layouter, err := buildElementTree(ctx, container)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem == nil {
		t.Fatal("expected non-nil element")
	}
	if elem.numChildren() != 2 {
		t.Errorf("expected 2 children, got %d", elem.numChildren())
	}
	if elem.child(0).Widget() != child1 {
		t.Errorf("first child widget not set correctly")
	}
	if elem.child(1).Widget().WidgetID() != ValueID("stateless") {
		t.Errorf("second child widget not set correctly")
	}

	if layouter == nil {
		t.Errorf("expected non-nil layouter for container widget")
	}
	children := slices.Collect(layouter.Children())
	if len(children) != 2 {
		t.Errorf("layouter should have 2 children, got %d", len(children))
	}
	if children[0] != layouter1 {
		t.Errorf("first child layouter not set correctly")
	}
	if children[1] != layouter2 {
		t.Errorf("second child layouter not set correctly")
	}
}

func TestBuildElementTree_ChildNoLayouter(t *testing.T) {
	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	childWidget := &mockWidget{ID: ValueID("child"), element: &ElementBase{}}
	container := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget {
			return childWidget
		})},
	}
	elem, layouter, err := buildElementTree(ctx, container)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elem.Widget() != container {
		t.Errorf("container widget not set correctly")
	}
	if elem.numChildren() != 1 {
		t.Errorf("expected 1 child, got %d", elem.numChildren())
	}
	if layouter == nil {
		t.Errorf("expected non-nil layouter for container widget")
	}
	children := slices.Collect(layouter.Children())
	if len(children) != 0 {
		t.Errorf("layouter should have 0 children, got %d", len(children))
	}
}

func TestUpdateElementTree_Reconcile(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &ElementBase{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}

	container1 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child2},
	}

	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	elem, layouter, err := buildElementTree(ctx, container1)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	container2 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			NewStatelessWidget(nil, func(ctx *Context) Widget { return child1 }),
			child2},
	}

	newElem, newLayouter, err := reconcileElementTree(ctx, elem, container2)
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
	if childWidget1, ok := newElem.child(0).Widget().(StatelessWidget); !ok {
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
	children := slices.Collect(newLayouter.Children())
	if len(children) != len(slices.Collect(layouter.Children())) || len(children) != 1 {
		t.Fatalf("expected layouter to have same number of children")
	}
	if children[0] != slices.Collect(layouter.Children())[0] {
		t.Fatalf("first child layouter not the same")
	}

	container3 := &mockContainer{
		Children: []Widget{
			NewStatelessWidget(ValueID("stateless"), func(ctx *Context) Widget { return child1 }),
			child2},
	}

	newElem2, newLayouter2, err := reconcileElementTree(ctx, newElem, container3)
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
	if newElem2.Widget() != container3 {
		t.Fatalf("new root element widget not set correctly")
	}
	if newElem2.numChildren() != 2 {
		t.Fatalf("expected 2 children, got %d", newElem2.numChildren())
	}
	if newElem2.child(0).Widget().WidgetID() != ValueID("stateless") {
		t.Fatal("first child widget not updated correctly")
	}
	if newElem2.child(1).Widget() != child2 {
		t.Fatalf("second child element not updated correctly")
	}
	children2 := slices.Collect(newLayouter2.Children())
	if len(children2) != 1 {
		t.Fatalf("expected layouter to have 1 child")
	}
	if children2[0] != slices.Collect(layouter.Children())[0] {
		t.Fatalf("first child layouter not the same")
	}
}

func TestUpdateElementTree_Reconcile_ID(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &ElementBase{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}
	child3 := &mockWidget{ID: ValueID("child3"), element: &ElementBase{}}

	container1 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			child1,
			NewStatefulWidget(nil, func(ctx *Context, updateState UpdateStateFunc) *WidgetState {
				return &WidgetState{
					Build: func() Widget { return child2 }}
			}),
			NewStatefulWidget(ValueID("no-change"), func(ctx *Context, updateState UpdateStateFunc) *WidgetState {
				return &WidgetState{
					Build: func() Widget { return child3 }}
			}),
		},
	}

	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	elem, layouter, err := buildElementTree(ctx, container1)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	if elem.numChildren() != 3 {
		t.Fatalf("expected 3 children, got %d", elem.numChildren())
	}

	child4 := &mockWidget{ID: ValueID("child1"), element: &ElementBase{}}
	child5 := &mockWidget{ID: ValueID("child5"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}
	container2 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			child4, // Match old #0, update in-place.
			NewStatefulWidget(ValueID("parent-of-child5"), func(ctx *Context, updateState UpdateStateFunc) *WidgetState { // No match, recrated
				return &WidgetState{
					Build: func() Widget { return child5 }}
			}),
			NewStatefulWidget(ValueID("no-change"), func(ctx *Context, updateState UpdateStateFunc) *WidgetState { // Match old #2, update in-place and createState will not be called.
				return &WidgetState{
					Build: func() Widget { panic("should not be called") }}
			}),
		},
	}

	newElem, newLayouter, err := reconcileElementTree(ctx, elem, container2)
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
	if newElem.numChildren() != 3 {
		t.Fatalf("expected 3 children, got %d", newElem.numChildren())
	}
	if id := newElem.child(0).Widget().WidgetID(); id != ValueID("child1") {
		t.Fatalf("expected first child element to be replaced, got %v", id)
	}
	if id := newElem.child(1).child(0).Widget().WidgetID(); id != ValueID("child5") {
		t.Fatalf("expected second child element to be replaced, got %v", id)
	}
	if id := newElem.child(2).child(0).Widget().WidgetID(); id != ValueID("child3") {
		t.Fatalf("expected third child element to be the same, got %v", id)
	}
}

func TestUpdateElementTree_Append(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &ElementBase{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}
	child3 := &mockWidget{ID: ValueID("child3"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}

	container1 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child2},
	}

	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	elem, layouter, err := buildElementTree(ctx, container1)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	container2 := &mockContainer{
		ID: ValueID("container"),
		Children: []Widget{
			child1, child2, child3},
	}

	newElem, newLayouter, err := reconcileElementTree(ctx, elem, container2)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}

	if newElem != elem {
		t.Fatalf("expected root element to be the same")
	}
	if newElem.numChildren() != 3 {
		t.Fatalf("expected 3 children, got %d", newElem.numChildren())
	}
	if newElem.child(0).Widget() != child1 || newElem.child(1).Widget() != child2 || newElem.child(2).Widget() != child3 {
		t.Fatalf("child elements not updated correctly")
	}

	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}
	children := slices.Collect(newLayouter.Children())
	if len(children) != 2 {
		t.Fatalf("expected layouter to have 2 children")
	}
	if children[0].Element().Widget() != child2 || children[1].Element().Widget() != child3 {
		t.Fatalf("child layouters not updated correctly")
	}
}

func TestUpdateElementTree_Remove(t *testing.T) {
	child1 := &mockWidget{ID: ValueID("child1"), element: &ElementBase{}}
	child2 := &mockWidget{ID: ValueID("child2"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}
	child3 := &mockWidget{ID: ValueID("child3"), element: &ElementBase{ElementLayouter: &mockLayouter{}}}

	container1 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child2, child3},
	}

	ctx := newMockContext(&AppConfig{Debug: &Debug{}})
	elem, layouter, err := buildElementTree(ctx, container1)
	if err != nil {
		t.Fatalf("unexpected error during build: %v", err)
	}

	container2 := &mockContainer{
		ID:       ValueID("container"),
		Children: []Widget{child1, child3},
	}

	newElem, newLayouter, err := reconcileElementTree(ctx, elem, container2)
	if err != nil {
		t.Fatalf("unexpected error during update: %v", err)
	}

	if newElem != elem {
		t.Fatalf("expected root element to be the same")
	}
	if newElem.numChildren() != 2 {
		t.Fatalf("expected 2 children, got %d", newElem.numChildren())
	}
	if newElem.child(0).Widget() != child1 || newElem.child(1).Widget() != child3 {
		t.Fatalf("child elements not updated correctly")
	}

	if newLayouter != layouter {
		t.Fatalf("expected root layouter to be the same")
	}
	children := slices.Collect(newLayouter.Children())
	if len(children) != 1 {
		t.Fatalf("expected layouter to have 1 child")
	}
	if children[0].Element().Widget() != child3 {
		t.Fatalf("child layouters not updated correctly")
	}
}
