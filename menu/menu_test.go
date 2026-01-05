package menu

import (
	"errors"
	"testing"

	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets/widgetstest"
)

// mockWidget is a test widget implementing [goui.Container]
// It can be used as a widget that is neither menu nor menu item, or
// as a container for menu or menu item.
type mockWidget struct {
	id    goui.ID
	child goui.Widget
}

func (w *mockWidget) WidgetID() goui.ID {
	return w.id
}

func (w *mockWidget) NumChildren() int {
	if w.child != nil {
		return 1
	}
	return 0
}

func (w *mockWidget) Child(n int) goui.Widget {
	if w.child == nil || n != 0 {
		panic("index out of range")
	}
	return w.child
}

func (w *mockWidget) Exclusive(goui.Container) {}

func (w *mockWidget) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementBase{}, nil
}

// mockStatelessWidget is a test widget implementing [goui.StatelessWidget].
// It can be used as a widget that is neither menu nor menu item, or
// as the parent of menu or menu item.
type mockStatelessWidget struct {
	goui.StatelessWidgetImpl
	id    goui.ID
	child goui.Widget
}

func (w *mockStatelessWidget) WidgetID() goui.ID {
	return w.id
}

func (w *mockStatelessWidget) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementBase{}, nil
}

func (w *mockStatelessWidget) Build(ctx *goui.Context) goui.Widget {
	return w.child
}

// mockStatefulWidget is a test widget implementing [goui.StatefulWidget].
// It can be used as a widget that is neither menu nor menu item, or
// as the parent of menu or menu item.
type mockStatefulWidget struct {
	goui.StatefulWidgetHelper
	id    goui.ID
	child goui.Widget
}

func (w *mockStatefulWidget) WidgetID() goui.ID {
	return w.id
}

func (w *mockStatefulWidget) CreateState(ctx *goui.StateContext) goui.State {
	return goui.NewState(ctx, func() goui.Widget {
		return w.child
	}, nil)
}

// TestBuildSuccess_BasicMenuStructure tests basic menu structure builds successfully
func TestBuildSuccess_BasicMenuStructure(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&Item{
				ID:    goui.ValueID("item1"),
				Title: "Item 1",
			},
			&Separator{ID: goui.ValueID("separator")},
			&Item{
				ID:    goui.ValueID("item2"),
				Title: "Item 2",
				Submenu: &Menu{
					ID: goui.ValueID("submenu"),
					Items: []goui.Widget{
						&Item{
							ID:    goui.ValueID("subitem"),
							Title: "Sub Item",
						},
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed for basic menu structure, got error: %v", err)
	}
}

// TestBuildSuccess_ItemWrappedByStateless tests Item wrapped by StatelessWidget builds successfully
func TestBuildSuccess_ItemWrappedByStateless(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockStatelessWidget{
				id: goui.ValueID("stateless"),
				child: &Item{
					ID:    goui.ValueID("item"),
					Title: "Item",
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Item is wrapped by StatelessWidget, got error: %v", err)
	}
}

// TestBuildSuccess_ItemWrappedByStateful tests Item wrapped by StatefulWidget builds successfully
func TestBuildSuccess_ItemWrappedByStateful(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockStatefulWidget{
				id: goui.ValueID("stateful"),
				child: &Item{
					ID:    goui.ValueID("item"),
					Title: "Item",
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Item is wrapped by StatefulWidget, got error: %v", err)
	}
}

// TestBuildSuccess_MenuWrappedByStateless tests Menu wrapped by StatelessWidget builds successfully
func TestBuildSuccess_MenuWrappedByStateless(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &mockStatelessWidget{
		id: goui.ValueID("stateless"),
		child: &Menu{
			ID:    goui.ValueID("submenu"),
			Items: []goui.Widget{},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Menu is wrapped by StatelessWidget, got error: %v", err)
	}
}

// TestBuildSuccess_SubmenuWrappedByStateless tests Submenu wrapped by StatelessWidget builds successfully
func TestBuildSuccess_SubmenuWrappedByStateless(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&Item{
				ID:    goui.ValueID("item"),
				Title: "Item",
				Submenu: &mockStatelessWidget{
					id: goui.ValueID("stateless"),
					child: &Menu{
						ID:    goui.ValueID("submenu"),
						Items: []goui.Widget{},
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Submenu is wrapped by StatelessWidget, got error: %v", err)
	}
}

// TestBuildSuccess_SubmenuWrappedByStateful tests Submenu wrapped by StatefulWidget builds successfully
func TestBuildSuccess_SubmenuWrappedByStateful(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&Item{
				ID:    goui.ValueID("item"),
				Title: "Item",
				Submenu: &mockStatefulWidget{
					id: goui.ValueID("stateful"),
					child: &Menu{
						ID:    goui.ValueID("submenu"),
						Items: []goui.Widget{},
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Submenu is wrapped by StatefulWidget, got error: %v", err)
	}
}

// TestBuildSuccess_NestedStatelessWidgets tests nested StatelessWidget wrappers build successfully
func TestBuildSuccess_NestedStatelessWidgets(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockStatelessWidget{
				id: goui.ValueID("stateless1"),
				child: &mockStatelessWidget{
					id: goui.ValueID("stateless2"),
					child: &Item{
						ID:    goui.ValueID("item"),
						Title: "Item",
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed with nested StatelessWidgets, got error: %v", err)
	}
}

// TestBuildSuccess_MixedWrappers tests mixed StatelessWidget and StatefulWidget wrappers build successfully
func TestBuildSuccess_MixedWrappers(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockStatefulWidget{
				id: goui.ValueID("stateful"),
				child: &Item{
					ID:    goui.ValueID("item"),
					Title: "Item",
					Submenu: &mockStatelessWidget{
						id: goui.ValueID("stateless"),
						child: &Menu{
							ID: goui.ValueID("submenu"),
							Items: []goui.Widget{
								&Item{
									ID:    goui.ValueID("subitem"),
									Title: "Sub Item",
								},
							},
						},
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed with mixed wrappers, got error: %v", err)
	}
}

// TestBuildSuccess_SeparatorWrappedByStateless tests Separator wrapped by StatelessWidget builds successfully
func TestBuildSuccess_SeparatorWrappedByStateless(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockStatelessWidget{
				id:    goui.ValueID("stateless"),
				child: &Separator{ID: goui.ValueID("separator")},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Separator is wrapped by StatelessWidget, got error: %v", err)
	}
}

// TestBuildFailure_ItemWrappedByContainer tests Item wrapped by Container fails to build
func TestBuildFailure_ItemWrappedByContainer(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockWidget{
				id: goui.ValueID("container"),
				child: &Item{
					ID:    goui.ValueID("item"),
					Title: "Item",
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err == nil {
		t.Fatalf("expected build to fail when Item is wrapped by Container (lookupNativeMenuParent blocked)")
	}
	if !errors.Is(err, ErrWrongParent) {
		t.Errorf("expected error to be ErrWrongParent, got: %v", err)
	}
}

// TestBuildFailure_SubmenuWrappedByContainer tests Submenu wrapped by Container fails to build
func TestBuildFailure_SubmenuWrappedByContainer(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&Item{
				ID:    goui.ValueID("item"),
				Title: "Item",
				Submenu: &mockWidget{
					id: goui.ValueID("container"),
					child: &Menu{
						ID:    goui.ValueID("submenu"),
						Items: []goui.Widget{},
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err == nil {
		t.Fatalf("expected build to fail when Submenu is wrapped by Container (lookupNativeItemParent blocked)")
	}
	if !errors.Is(err, ErrWrongParent) {
		t.Errorf("expected error to be ErrWrongParent, got: %v", err)
	}
}

// TestBuildFailure_SeparatorWrappedByContainer tests Separator wrapped by Container fails to build
func TestBuildFailure_SeparatorWrappedByContainer(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockWidget{
				id:    goui.ValueID("container"),
				child: &Separator{ID: goui.ValueID("separator")},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err == nil {
		t.Fatalf("expected build to fail when Separator is wrapped by Container (lookupNativeMenuParent blocked)")
	}
	if !errors.Is(err, ErrWrongParent) {
		t.Errorf("expected error to be ErrWrongParent, got: %v", err)
	}
}

// TestBuildFailure_ContainerBeforeStateless tests Container before StatelessWidget fails to build
func TestBuildFailure_ContainerBeforeStateless(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.Widget{
			&mockWidget{
				id: goui.ValueID("container"),
				child: &mockStatelessWidget{
					id: goui.ValueID("stateless"),
					child: &Item{
						ID:    goui.ValueID("item"),
						Title: "Item",
					},
				},
			},
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err == nil {
		t.Fatalf("expected build to fail when Container is between Menu and Item, even with StatelessWidget after")
	}
	if !errors.Is(err, ErrWrongParent) {
		t.Errorf("expected error to be ErrWrongParent, got: %v", err)
	}
}
