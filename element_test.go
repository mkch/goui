package goui

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/mkch/gg"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type mockWidget struct {
	ID          ID
	createError error
	element     Element
}

func (w *mockWidget) WidgetID() ID {
	return w.ID
}

func (w *mockWidget) CreateElement(ctx *Context, parent Element) (Element, error) {
	if w.createError != nil {
		return nil, w.createError
	}
	return w.element, nil
}

func (*mockWidget) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

func TestBuildElementTree_CreateElementError(t *testing.T) {
	runContextMock(func(ctx *Context) {
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
	}, nil)

}

func TestBuildElementTree_SimpleWidget(t *testing.T) {
	runContextMock(func(ctx *Context) {
		widget := &mockWidget{ID: ValueID("test"), element: &ElementHelper{}}

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
		if elem.NumChildren() != 0 {
			t.Errorf("expected 0 children, got %d", elem.NumChildren())
		}
		if layouter != nil {
			t.Errorf("expected nil layouter for simple widget, got %v", layouter)
		}
	}, &AppConfig{Debug: &Debug{}})

}

type mockLayouter struct {
	LayouterHelper
}

func (l *mockLayouter) Layout(ctx *Context, constraints metrics.Constraints) (metrics.Size, error) {
	return metrics.Size{Width: 100, Height: 100}, nil
}

func (l *mockLayouter) PositionAt(ctx *Context, pt metrics.Point) error {
	return nil
}

func TestBuildElementTree_WidgetWithLayouter(t *testing.T) {
	runContextMock(func(ctx *Context) {
		mockLayouter := &mockLayouter{}
		mockElement := &ElementHelper{
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
	}, &AppConfig{Debug: &Debug{}})

}

func TestBuildElementTree_StatefulWidget(t *testing.T) {
	runContextMock(func(ctx *Context) {
		mockLayouter := &mockLayouter{}
		mockElement := &ElementHelper{
			ElementLayouter: mockLayouter,
		}
		childWidget := &mockWidget{ID: ValueID("child"), element: mockElement}

		widget := NewStatefulWidget(
			ValueID("stateful"),
			func(ctx *StateContext) State {
				return NewState(ctx, func() Widget { return childWidget }, nil)
			},
		)

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
		if elem.NumChildren() != 1 {
			t.Errorf("expected 1 child, got %d", elem.NumChildren())
		}
		if elem.Child(0).Widget().WidgetID() != childWidget.WidgetID() {
			t.Errorf("child widget not set correctly")
		}
		if layouter != mockLayouter {
			t.Errorf("wrong layouter returned")
		}
	}, &AppConfig{Debug: &Debug{}})

}

func TestBuildElementTree_StatelessWidget(t *testing.T) {
	runContextMock(func(ctx *Context) {
		mockLayouter := &mockLayouter{}
		mockElement := &ElementHelper{
			ElementLayouter: mockLayouter,
		}
		childWidget := &mockWidget{ID: ValueID("child"), element: mockElement}

		widget := NewStatelessWidget(
			ValueID("stateless"),
			func(ctx *Context) Widget {
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
		if elem.NumChildren() != 1 {
			t.Errorf("expected 1 child, got %d", elem.NumChildren())
		}
		if elem.Child(0).Widget().WidgetID() != childWidget.WidgetID() {
			t.Errorf("child widget not set correctly")
		}
		if layouter != mockLayouter {
			t.Errorf("wrong layouter returned")
		}
	}, &AppConfig{Debug: &Debug{}})
}

type mockContainer struct {
	ID       ID
	Children []Widget
}

func (c *mockContainer) WidgetID() ID {
	return c.ID
}

func (c *mockContainer) CreateElement(ctx *Context, parent Element) (Element, error) {
	return &ElementHelper{ElementLayouter: &mockLayouter{}}, nil
}

func (c *mockContainer) NumChildren() int {
	return len(c.Children)
}

func (c *mockContainer) Child(n int) AbstractWidget {
	return c.Children[n]
}

func (*mockContainer) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

func (*mockContainer) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

func TestBuildElementTree_Container(t *testing.T) {
	runContextMock(func(ctx *Context) {
		layouter1 := &mockLayouter{}
		child1 := &mockWidget{ID: ValueID("child1"), element: &ElementHelper{ElementLayouter: layouter1}}
		layouter2 := &mockLayouter{}
		child2 := &mockWidget{ID: ValueID("child2"), element: &ElementHelper{ElementLayouter: layouter2}}

		container := &mockContainer{
			ID: ValueID("container"),
			Children: []Widget{
				child1,
				NewStatelessWidget(
					ValueID("stateless"),
					func(ctx *Context) Widget {
						return child2
					},
				)},
		}

		elem, layouter, err := buildElementTree(ctx, container)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if elem == nil {
			t.Fatal("expected non-nil element")
		}
		if elem.NumChildren() != 2 {
			t.Errorf("expected 2 children, got %d", elem.NumChildren())
		}
		if elem.Child(0).Widget() != child1 {
			t.Errorf("first child widget not set correctly")
		}
		if elem.Child(1).Widget().WidgetID() != ValueID("stateless") {
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
	}, &AppConfig{Debug: &Debug{}})
}

func TestBuildElementTree_ChildNoLayouter(t *testing.T) {
	runContextMock(func(ctx *Context) {
		childWidget := &mockWidget{ID: ValueID("child"), element: &ElementHelper{}}
		container := &mockContainer{
			ID: ValueID("container"),
			Children: []Widget{NewStatelessWidget(
				ValueID("stateless"),
				func(ctx *Context) Widget {
					return childWidget
				},
			)},
		}
		elem, layouter, err := buildElementTree(ctx, container)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if elem.Widget() != container {
			t.Errorf("container widget not set correctly")
		}
		if elem.NumChildren() != 1 {
			t.Errorf("expected 1 child, got %d", elem.NumChildren())
		}
		if layouter == nil {
			t.Errorf("expected non-nil layouter for container widget")
		}
		children := slices.Collect(layouter.Children())
		if len(children) != 0 {
			t.Errorf("layouter should have 0 children, got %d", len(children))
		}
	}, &AppConfig{Debug: &Debug{}})
}

func TestUpdateElementTree_Reconcile(t *testing.T) {
	runContextMock(func(ctx *Context) {
		child1 := &mockWidget{ID: ValueID("child1"), element: &ElementHelper{}}
		child2 := &mockWidget{ID: ValueID("child2"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}

		container1 := &mockContainer{
			ID:       ValueID("container"),
			Children: []Widget{child1, child2},
		}

		elem, layouter, err := buildElementTree(ctx, container1)
		if err != nil {
			t.Fatalf("unexpected error during build: %v", err)
		}

		container2 := &mockContainer{
			ID: ValueID("container"),
			Children: []Widget{
				StatelessWidgetFunc(func(ctx *Context) Widget {
					return child1
				}),
				child2},
		}

		newElem, err := reconcileElementTreeImpl(ctx, elem, container2)
		if err != nil {
			t.Fatalf("unexpected error during update: %v", err)
		}
		newLayouter := layouterTree(newElem)

		// The root element and layouter should be the same.
		if newElem != elem {
			t.Fatalf("expected root element to be the same")
		}
		if newLayouter != layouter {
			t.Fatalf("expected root layouter to be the same")
		}
		// The first child element should be replaced.
		if newElem.NumChildren() != 2 {
			t.Fatalf("expected 2 children, got %d", newElem.NumChildren())
		}
		if childWidget1, ok := newElem.Child(0).Widget().(StatelessWidget); !ok {
			t.Fatal("expected first child to be a StatelessWidget")
		} else if childWidget1.Build(ctx) != child1 {
			t.Fatal("first child widget not updated correctly")
		}
		// The second child element should be the same.
		if child2 := newElem.Child(1); child2 != child2 {
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
				NewStatelessWidget(
					ValueID("stateless"),
					func(ctx *Context) Widget {
						return child1
					},
				),
				child2,
			},
		}

		newElem2, err := reconcileElementTreeImpl(ctx, newElem, container3)
		if err != nil {
			t.Fatalf("unexpected error during update: %v", err)
		}
		newLayouter2 := layouterTree(newElem2)

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
		if newElem2.NumChildren() != 2 {
			t.Fatalf("expected 2 children, got %d", newElem2.NumChildren())
		}
		if newElem2.Child(0).Widget().WidgetID() != ValueID("stateless") {
			t.Fatal("first child widget not updated correctly")
		}
		if newElem2.Child(1).Widget() != child2 {
			t.Fatalf("second child element not updated correctly")
		}
		children2 := slices.Collect(newLayouter2.Children())
		if len(children2) != 1 {
			t.Fatalf("expected layouter to have 1 child")
		}
		if children2[0] != slices.Collect(layouter.Children())[0] {
			t.Fatalf("first child layouter not the same")
		}
	}, &AppConfig{Debug: &Debug{}})
}

func TestUpdateElementTree_Reconcile_ID(t *testing.T) {
	runContextMock(func(ctx *Context) {
		child1 := &mockWidget{ID: ValueID("child1"), element: &ElementHelper{}}
		child2 := &mockWidget{ID: ValueID("child2"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}
		child3 := &mockWidget{ID: ValueID("child3"), element: &ElementHelper{}}

		container1 := &mockContainer{
			ID: ValueID("container"),
			Children: []Widget{
				child1,
				NewStatefulWidget(
					nil,
					func(ctx *StateContext) State {
						return NewState(ctx, func() Widget { return child2 }, nil)
					},
				),
				NewStatefulWidget(
					ValueID("no-change"),
					func(ctx *StateContext) State {
						return NewState(ctx, func() Widget { return child3 }, nil)
					},
				),
			},
		}

		elem, layouter, err := buildElementTree(ctx, container1)
		if err != nil {
			t.Fatalf("unexpected error during build: %v", err)
		}

		if elem.NumChildren() != 3 {
			t.Fatalf("expected 3 children, got %d", elem.NumChildren())
		}

		child4 := &mockWidget{ID: ValueID("child1"), element: &ElementHelper{}}
		child5 := &mockWidget{ID: ValueID("child5"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}
		container2 := &mockContainer{
			ID: ValueID("container"),
			Children: []Widget{
				child4, // Match old #0, update in-place.
				NewStatefulWidget(
					ValueID("parent-of-child5"),
					func(ctx *StateContext) State { // No match, recrated
						return NewState(ctx, func() Widget { return child5 }, nil)
					},
				),
				NewStatefulWidget(
					ValueID("no-change"),
					func(ctx *StateContext) State { // Match old #2, update in-place and createState will not be called.
						return NewState(ctx, func() Widget { return child3 }, nil)
					},
				),
			},
		}

		newElem, err := reconcileElementTreeImpl(ctx, elem, container2)
		if err != nil {
			t.Fatalf("unexpected error during update: %v", err)
		}
		newLayouter := layouterTree(newElem)

		// The root element and layouter should be the same.
		if newElem != elem {
			t.Fatalf("expected root element to be the same")
		}
		if newLayouter != layouter {
			t.Fatalf("expected root layouter to be the same")
		}

		// The child elements should be replaced.
		if newElem.NumChildren() != 3 {
			t.Fatalf("expected 3 children, got %d", newElem.NumChildren())
		}
		if id := newElem.Child(0).Widget().WidgetID(); id != ValueID("child1") {
			t.Fatalf("expected first child element to be replaced, got %v", id)
		}
		if id := newElem.Child(1).Child(0).Widget().WidgetID(); id != ValueID("child5") {
			t.Fatalf("expected second child element to be replaced, got %v", id)
		}
		if id := newElem.Child(2).Child(0).Widget().WidgetID(); id != ValueID("child3") {
			t.Fatalf("expected third child element to be the same, got %v", id)
		}
	}, &AppConfig{Debug: &Debug{}})
}

func TestUpdateElementTree_Append(t *testing.T) {
	runContextMock(func(ctx *Context) {
		child1 := &mockWidget{ID: ValueID("child1"), element: &ElementHelper{}}
		child2 := &mockWidget{ID: ValueID("child2"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}
		child3 := &mockWidget{ID: ValueID("child3"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}

		container1 := &mockContainer{
			ID:       ValueID("container"),
			Children: []Widget{child1, child2},
		}
		elem, layouter, err := buildElementTree(ctx, container1)
		if err != nil {
			t.Fatalf("unexpected error during build: %v", err)
		}

		container2 := &mockContainer{
			ID: ValueID("container"),
			Children: []Widget{
				child1, child2, child3},
		}

		newElem, err := reconcileElementTreeImpl(ctx, elem, container2)
		if err != nil {
			t.Fatalf("unexpected error during update: %v", err)
		}
		newLayouter := layouterTree(newElem)

		if newElem != elem {
			t.Fatalf("expected root element to be the same")
		}
		if newElem.NumChildren() != 3 {
			t.Fatalf("expected 3 children, got %d", newElem.NumChildren())
		}
		if newElem.Child(0).Widget() != child1 || newElem.Child(1).Widget() != child2 || newElem.Child(2).Widget() != child3 {
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
	}, &AppConfig{Debug: &Debug{}})
}

func TestUpdateElementTree_Remove(t *testing.T) {
	runContextMock(func(ctx *Context) {
		child1 := &mockWidget{ID: ValueID("child1"), element: &ElementHelper{}}
		child2 := &mockWidget{ID: ValueID("child2"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}
		child3 := &mockWidget{ID: ValueID("child3"), element: &ElementHelper{ElementLayouter: &mockLayouter{}}}

		container1 := &mockContainer{
			ID:       ValueID("container"),
			Children: []Widget{child1, child2, child3},
		}
		elem, layouter, err := buildElementTree(ctx, container1)
		if err != nil {
			t.Fatalf("unexpected error during build: %v", err)
		}

		container2 := &mockContainer{
			ID:       ValueID("container"),
			Children: []Widget{child1, child3},
		}

		newElem, err := reconcileElementTreeImpl(ctx, elem, container2)
		if err != nil {
			t.Fatalf("unexpected error during update: %v", err)
		}
		newLayouter := layouterTree(newElem)

		if newElem != elem {
			t.Fatalf("expected root element to be the same")
		}
		if newElem.NumChildren() != 2 {
			t.Fatalf("expected 2 children, got %d", newElem.NumChildren())
		}
		if newElem.Child(0).Widget() != child1 || newElem.Child(1).Widget() != child3 {
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
	}, &AppConfig{Debug: &Debug{}})
}

type mockMenu struct {
	ID     ID
	Widget Widget
}

func (m *mockMenu) WidgetID() ID {
	return m.ID
}

func (m *mockMenu) CreateElement(ctx *Context, parent Element) (Element, error) {
	var elem mockMenuElement
	elem.Handle = 1001
	return &elem, nil
}

func (m *mockMenu) NumChildren() int {
	return gg.If(m.Widget != nil, 1, 0)
}

func (m *mockMenu) Child(n int) Widget {
	if n != 0 || m.Widget == nil {
		panic("index out of range")
	}
	return m.Widget
}

func (*mockMenu) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

func (*mockMenu) ExclusiveType(marker.TypeMenu) { /*Nop*/ }

type mockMenuElement struct {
	ElementHelper
	Handle native.Handle
}

func (e *mockMenuElement) NativeMenu() native.Handle {
	return e.Handle
}

type mockControl struct {
	ID ID
}

func (c *mockControl) WidgetID() ID {
	return c.ID
}

func (c *mockControl) CreateElement(ctx *Context, parent Element) (Element, error) {
	nativeParent, err := WidgetNativeParent(ctx, parent)
	if err != nil {
		return nil, fmt.Errorf("Create mock control failed: %w", err)
	}
	var elem mockControlElement
	elem.Handle = 1
	elem.nativeParent = nativeParent
	return &elem, nil
}

type mockControlElement struct {
	ControlElementHelper
	nativeParent native.Handle
}
