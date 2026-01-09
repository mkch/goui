package menu

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
)

// StatelessMenu is a stateless wrapper for a [Menu].
type StatelessMenu interface {
	goui.AbstractStatelessWidget
	ExclusiveType(marker.TypeMenu)
}

// statelessMenu is a [StatelessMenu] implementation.
type statelessMenu struct {
	goui.StatelessHelper
}

func (*statelessMenu) ExclusiveType(marker.TypeMenu) { /*Nop*/ }

func NewStatelessMenu(ID goui.ID, builder func(ctx *goui.Context) goui.Menu) StatelessMenu {
	return &statelessMenu{
		StatelessHelper: goui.StatelessHelper{
			ID:      ID,
			Builder: func(ctx *goui.Context) goui.AbstractWidget { return builder(ctx) },
		},
	}
}

// StatelessMenuFunc is a function type that implements [StatelessMenu].
type StatelessMenuFunc func(ctx *goui.Context) goui.Menu

// WidgetID implements [StatelessMenu.WidgetID] and always returns nil.
func (f StatelessMenuFunc) WidgetID() goui.ID {
	return nil
}

// CreateElement implements [StatelessMenu.CreateElement].
func (f StatelessMenuFunc) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementHelper{}, nil
}

// Build implements [StatelessMenu.Build].
// It calls f(ctx).
func (f StatelessMenuFunc) Build(ctx *goui.Context) goui.AbstractWidget {
	return f(ctx)
}

func (f StatelessMenuFunc) ExclusiveType(marker.TypeMenu)      { /*Nop*/ }
func (f StatelessMenuFunc) ExclusiveKind(marker.KindStateless) { /*Nop*/ }

type StatelessItem interface {
	goui.AbstractStatelessWidget
	ExclusiveType(marker.TypeMenuItem)
}

type statelessItem struct {
	goui.StatelessHelper
}

func (*statelessItem) ExclusiveType(marker.TypeMenuItem) { /*Nop*/ }

func NewStatelessItem(ID goui.ID, builder func(ctx *goui.Context) goui.MenuItem) StatelessItem {
	return &statelessItem{
		StatelessHelper: goui.StatelessHelper{
			ID:      ID,
			Builder: func(ctx *goui.Context) goui.AbstractWidget { return builder(ctx) },
		},
	}
}

// StatelessItemFunc is a function type that implements [StatelessItem].
type StatelessItemFunc func(ctx *goui.Context) goui.MenuItem

// WidgetID implements [StatelessItem.WidgetID] and always returns nil.
func (f StatelessItemFunc) WidgetID() goui.ID {
	return nil
}

// CreateElement implements [StatelessItem.CreateElement].
func (f StatelessItemFunc) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementHelper{}, nil
}

// Build implements [StatelessMenu.Build].
// It calls f(ctx).
func (f StatelessItemFunc) Build(ctx *goui.Context) goui.AbstractWidget {
	return f(ctx)
}

func (f StatelessItemFunc) ExclusiveType(marker.TypeMenuItem)  { /*Nop*/ }
func (f StatelessItemFunc) ExclusiveKind(marker.KindStateless) { /*Nop*/ }
