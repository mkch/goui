package menu

import (
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

func (w *mockWidget) Child(n int) goui.WidgetBase {
	if w.child == nil || n != 0 {
		panic("index out of range")
	}
	return w.child
}

func (w *mockWidget) Exclusive(goui.ContainerBase) {}

func (w *mockWidget) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementBase{}, nil
}

// TestBuildSuccess_BasicMenuStructure tests basic menu structure builds successfully
func TestBuildSuccess_BasicMenuStructure(t *testing.T) {
	ctx := widgetstest.NewContext()

	menu := &Menu{
		ID: goui.ValueID("menu"),
		Items: []goui.MenuItem{
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
					Items: []goui.MenuItem{
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
		Items: []goui.MenuItem{
			NewStatelessItem(
				goui.ValueID("stateless"),
				func(ctx *goui.Context) goui.MenuItem {
					return &Item{
						ID:    goui.ValueID("item"),
						Title: "Item",
					}
				},
			),
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
		Items: []goui.MenuItem{
			NewStatefulItem(
				goui.ValueID("stateful"),
				func(ctx *goui.StateContext) ItemState {
					return NewItemState(ctx, func() goui.MenuItem {
						return &Item{
							ID:    goui.ValueID("item"),
							Title: "Item",
						}
					}, nil)
				},
			),
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

	menu := NewStatelessMenu(
		goui.ValueID("stateless"),
		func(ctx *goui.Context) goui.Menu {
			return &Menu{
				ID: goui.ValueID("submenu"),
			}
		},
	)

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
		Items: []goui.MenuItem{
			&Item{
				ID:    goui.ValueID("item"),
				Title: "Item",
				Submenu: NewStatelessMenu(
					goui.ValueID("stateless"),
					func(ctx *goui.Context) goui.Menu {
						return &Menu{
							ID:    goui.ValueID("submenu"),
							Items: []goui.MenuItem{},
						}
					},
				),
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
		Items: []goui.MenuItem{
			&Item{
				ID:    goui.ValueID("item"),
				Title: "Item",
				Submenu: NewStatefulMenu(
					goui.ValueID("stateful"),
					func(ctx *goui.StateContext) MenuState {
						return NewMenuState(ctx, func() goui.Menu {
							return &Menu{
								ID:    goui.ValueID("submenu"),
								Items: []goui.MenuItem{},
							}
						}, nil)
					},
				),
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
		Items: []goui.MenuItem{
			NewStatelessItem(
				goui.ValueID("stateless1"),
				func(ctx *goui.Context) goui.MenuItem {
					return NewStatelessItem(
						goui.ValueID("stateless2"),
						func(ctx *goui.Context) goui.MenuItem {
							return &Item{
								ID:    goui.ValueID("item"),
								Title: "Item",
							}
						},
					)
				},
			),
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
		Items: []goui.MenuItem{
			NewStatefulItem(
				goui.ValueID("stateful"),
				func(ctx *goui.StateContext) ItemState {
					return NewItemState(ctx, func() goui.MenuItem {
						return &Item{
							ID:    goui.ValueID("item"),
							Title: "Item",
							Submenu: NewStatelessMenu(
								goui.ValueID("stateless"),
								func(ctx *goui.Context) goui.Menu {
									return &Menu{
										ID: goui.ValueID("submenu"),
										Items: []goui.MenuItem{
											&Item{
												ID:    goui.ValueID("subitem"),
												Title: "Sub Item",
											},
										},
									}
								},
							),
						}
					}, nil)
				},
			),
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
		Items: []goui.MenuItem{
			NewStatelessItem(
				goui.ValueID("stateless"),
				func(ctx *goui.Context) goui.MenuItem {
					return &Separator{ID: goui.ValueID("separator")}
				},
			),
		},
	}

	_, _, err := widgetstest.BuildElementTree(ctx, menu, nil)
	if err != nil {
		t.Errorf("expected build to succeed when Separator is wrapped by StatelessWidget, got error: %v", err)
	}
}
