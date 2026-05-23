package menu

import (
	"testing"

	"github.com/mkch/goui"
	"github.com/mkch/goui/gouitest"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/native"
	"github.com/mkch/goui/native/mock"
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

func (w *mockWidget) Child(n int) goui.Component {
	if w.child == nil || n != 0 {
		panic("index out of range")
	}
	return w.child
}

func (*mockWidget) ExclusiveKind(marker.KindContainer) {}

func (w *mockWidget) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementHelper{}, nil
}

// TestBuildSuccess_BasicMenuStructure tests basic menu structure builds successfully
func TestBuildSuccess_BasicMenuStructure(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed for basic menu structure, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 3 {
			t.Errorf("expected main menu to have 3 items, got %d", len(mainMenuItems))
		}
		item1Title, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item1: %v", err)
		}
		if item1Title != "Item 1" {
			t.Errorf("expected item1 title to be 'Item 1', got '%s'", item1Title)
		}
		item2IsSeparator, err := os.Debug_MenuItemIsSeparator(mainMenuItems[1])
		if err != nil {
			t.Errorf("failed to check if second item is separator: %v", err)
		}
		if !item2IsSeparator {
			t.Errorf("expected second item to be a separator, got non-separator")
		}
		item3Title, err := os.Debug_MenuItemTitle(mainMenuItems[2])
		if err != nil {
			t.Errorf("failed to get title of item2: %v", err)
		}
		if item3Title != "Item 2" {
			t.Errorf("expected item2 title to be 'Item 2', got '%s'", item3Title)
		}

		subMenu, err := os.Debug_MenuItemSubMenu(mainMenuItems[2])
		if err != nil {
			t.Errorf("failed to get submenu of item2: %v", err)
		}
		if subMenu == nil {
			t.Errorf("expected item2 to have a submenu, got nil")
		}
		subMenuItems, err := os.Debug_MenuItems(subMenu)
		if err != nil {
			t.Errorf("failed to get menu items of submenu: %v", err)
		}
		if len(subMenuItems) != 1 {
			t.Errorf("expected submenu to have 1 item, got %d", len(subMenuItems))
		}
		subItemTitle, err := os.Debug_MenuItemTitle(subMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of subitem: %v", err)
		}
		if subItemTitle != "Sub Item" {
			t.Errorf("expected subitem title to be 'Sub Item', got '%s'", subItemTitle)
		}

		item1Title, err = os.Debug_MenuItemTitle(nativeHandles[goui.ValueID("item1")])
		if err != nil {
			t.Errorf("failed to get title of item1 by its own handle: %v", err)
		}
		if item1Title != "Item 1" {
			t.Errorf("expected item1 title to be 'Item 1' when accessed by its own handle, got '%s'", item1Title)
		}

		item2Title, err := os.Debug_MenuItemTitle(nativeHandles[goui.ValueID("item2")])
		if err != nil {
			t.Errorf("failed to get title of item2 by its ID: %v", err)
		}
		if item2Title != "Item 2" {
			t.Errorf("expected item2 title to be 'Item 2' when accessed by its ID, got '%s'", item2Title)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_ItemWrappedByStateless tests Item wrapped by StatelessWidget builds successfully
func TestBuildSuccess_ItemWrappedByStateless(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed when Item is wrapped by StatelessWidget, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		itemTitle, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item', got '%s'", itemTitle)
		}

		itemTitle, err = os.Debug_MenuItemTitle(nativeHandles[goui.ValueID("item")])
		if err != nil {
			t.Errorf("failed to get title of item by its own handle: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item' when accessed by its own handle, got '%s'", itemTitle)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_ItemWrappedByStateful tests Item wrapped by StatefulWidget builds successfully
func TestBuildSuccess_ItemWrappedByStateful(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		menu := &Menu{
			ID: goui.ValueID("menu"),
			Items: []goui.MenuItem{
				StatefulItemFunc(
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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed when Item is wrapped by StatefulWidget, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		itemTitle, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item', got '%s'", itemTitle)
		}

		itemTitle, err = os.Debug_MenuItemTitle(nativeHandles[goui.ValueID("item")])
		if err != nil {
			t.Errorf("failed to get title of item by its own handle: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item' when accessed by its own handle, got '%s'", itemTitle)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_MenuWrappedByStateless tests Menu wrapped by StatelessWidget builds successfully
func TestBuildSuccess_MenuWrappedByStateless(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		menu := NewStatelessMenu(
			goui.ValueID("stateless"),
			func(ctx *goui.Context) goui.Menu {
				return &Menu{
					ID: goui.ValueID("submenu"),
				}
			},
		)

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed when Menu is wrapped by StatelessWidget, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("submenu")]
		if mainMenu == nil {
			t.Errorf("expected submenu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of submenu: %v", err)
		}
		if len(mainMenuItems) != 0 {
			t.Errorf("expected submenu to have 0 items, got %d", len(mainMenuItems))
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_SubmenuWrappedByStateless tests Submenu wrapped by StatelessWidget builds successfully
func TestBuildSuccess_SubmenuWrappedByStateless(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed when Submenu is wrapped by StatelessWidget, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		itemTitle, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item', got '%s'", itemTitle)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_SubmenuWrappedByStateful tests Submenu wrapped by StatefulWidget builds successfully
func TestBuildSuccess_SubmenuWrappedByStateful(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		menu := &Menu{
			ID: goui.ValueID(0),
			Items: []goui.MenuItem{
				&Item{
					ID:    goui.ValueID("item"),
					Title: "Item",
					Submenu: StatefulMenuFunc(
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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed when Submenu is wrapped by StatefulWidget, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID(0)]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		itemTitle, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item', got '%s'", itemTitle)
		}
		subMenu, err := os.Debug_MenuItemSubMenu(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get submenu of item: %v", err)
		}
		if subMenu == nil {
			t.Errorf("expected item to have a submenu, got nil")
		}
		subMenuItems, err := os.Debug_MenuItems(subMenu)
		if err != nil {
			t.Errorf("failed to get menu items of submenu: %v", err)
		}
		if len(subMenuItems) != 0 {
			t.Errorf("expected submenu to have 0 items, got %d", len(subMenuItems))
		}

		itemTitle, err = os.Debug_MenuItemTitle(nativeHandles[goui.ValueID("item")])
		if err != nil {
			t.Errorf("failed to get title of item by its own handle: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item' when accessed by its own handle, got '%s'", itemTitle)
		}
		subMenuItems, err = os.Debug_MenuItems(nativeHandles[goui.ValueID("submenu")])
		if err != nil {
			t.Errorf("failed to get submenu by its own handle: %v", err)
		}
		if len(subMenuItems) != 0 {
			t.Errorf("expected submenu to have 0 items when accessed by its own handle, got %d", len(subMenuItems))
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_NestedStatelessWidgets tests nested StatelessWidget wrappers build successfully
func TestBuildSuccess_NestedStatelessWidgets(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed with nested StatelessWidgets, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		itemTitle, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item', got '%s'", itemTitle)
		}

		itemTitle, err = os.Debug_MenuItemTitle(nativeHandles[goui.ValueID("item")])
		if err != nil {
			t.Errorf("failed to get title of item by its own handle: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item' when accessed by its own handle, got '%s'", itemTitle)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_MixedWrappers tests mixed StatelessWidget and StatefulWidget wrappers build successfully
func TestBuildSuccess_MixedWrappers(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed with mixed wrappers, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		itemTitle, err := os.Debug_MenuItemTitle(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of item: %v", err)
		}
		if itemTitle != "Item" {
			t.Errorf("expected item title to be 'Item', got '%s'", itemTitle)
		}
		subMenu, err := os.Debug_MenuItemSubMenu(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to get submenu of item: %v", err)
		}
		if subMenu == nil {
			t.Errorf("expected item to have a submenu, got nil")
		}
		subMenuItems, err := os.Debug_MenuItems(subMenu)
		if err != nil {
			t.Errorf("failed to get menu items of submenu: %v", err)
		}
		if len(subMenuItems) != 1 {
			t.Errorf("expected submenu to have 1 item, got %d", len(subMenuItems))
		}
		subItemTitle, err := os.Debug_MenuItemTitle(subMenuItems[0])
		if err != nil {
			t.Errorf("failed to get title of subitem: %v", err)
		}
		if subItemTitle != "Sub Item" {
			t.Errorf("expected subitem title to be 'Sub Item', got '%s'", subItemTitle)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})
}

// TestBuildSuccess_SeparatorWrappedByStateless tests Separator wrapped by StatelessWidget builds successfully
func TestBuildSuccess_SeparatorWrappedByStateless(t *testing.T) {
	nativeHandles := make(map[goui.ID]native.Handle)

	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

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

		_, _, err := gouitest.BuildElementTree(ctx, menu, nil)
		if err != nil {
			t.Errorf("expected build to succeed when Separator is wrapped by StatelessWidget, got error: %v", err)
		}

		// Check native menu structure
		mainMenu := nativeHandles[goui.ValueID("menu")]
		if mainMenu == nil {
			t.Errorf("expected main menu to be created, got nil handle")
		}
		os := ctx.App().OS().(*mock.OS)
		mainMenuItems, err := os.Debug_MenuItems(mainMenu)
		if err != nil {
			t.Errorf("failed to get menu items of main menu: %v", err)
		}
		if len(mainMenuItems) != 1 {
			t.Errorf("expected main menu to have 1 item, got %d", len(mainMenuItems))
		}
		isSeparator, err := os.Debug_MenuItemIsSeparator(mainMenuItems[0])
		if err != nil {
			t.Errorf("failed to check if menu item is separator: %v", err)
		}
		if !isSeparator {
			t.Errorf("expected menu item to be a separator, got non-separator")
		}
	}, &goui.AppConfig{Debug: &goui.Debug{
		NativeHandleRecords: nativeHandles,
	}})

}
