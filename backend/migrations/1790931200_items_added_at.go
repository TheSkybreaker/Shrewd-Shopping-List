package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// added_at orders the items to buy. An item put back from the cart goes to the end, which its
// creation date cannot express because putting it back is an update.
func init() {
	m.Register(func(app core.App) error {
		items, err := app.FindCollectionByNameOrId("items")
		if err != nil {
			return err
		}

		items.Fields.Add(&core.DateField{Name: "added_at"})

		return app.Save(items)
	}, func(app core.App) error {
		items, err := app.FindCollectionByNameOrId("items")
		if err != nil {
			return err
		}

		items.Fields.RemoveByName("added_at")

		return app.Save(items)
	})
}
