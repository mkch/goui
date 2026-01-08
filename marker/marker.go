// Package marker defines marker type used by ExclusiveType and ExclusiveKind methods.
//
// Exclusive marker methods are used to declare mutually exclusive types or kinds.
// For example, if a type has method
//
//	ExclusiveType(marker.TypeWidget)
//
// it cannot have
//
//	ExclusiveType(marker.TypeMenu)
//
// or
//
//	ExclusiveType(marker.TypeMenuItem).
//
// This means a type can be either a Widget, a Menu, or a MenuItem, but not more than one of these.
// The kind dimension works similarly with KindStateful, KindStateless, and KindContainer.
//
// This mechanism is not perfect.For example, it cannot prevent a type from having both
//
//	ExclusiveType(marker.TypeMenuItem)
//	ExclusiveKind(marker.Container)
//
// which makes no sense, because it is not realistic to implement a menu item that contains other things.
// But it is sufficient for most practical use cases.
package marker

// TypeWidget is a marker type for window widgets.
// Used as parameter type of ExclusiveType method.
type TypeWidget struct{}

// TypeMenu is a marker type for menus.
// Used as parameter type of ExclusiveType method.
type TypeMenu struct{}

// TypeMenuItem is a marker type for menu items.
// Used as parameter type of ExclusiveType method.
type TypeMenuItem struct{}

// KindStateful is a marker type for stateful widgets, menus or menu items.
// Used as parameter type of ExclusiveKind method.
type KindStateful struct{}

// KindStateless is a marker type for stateless widgets, menus or menu items.
// Used as parameter type of ExclusiveKind method.
type KindStateless struct{}

// KindContainer is a marker type for container widgets.
// Used as parameter type of ExclusiveKind method.
type KindContainer struct{}
