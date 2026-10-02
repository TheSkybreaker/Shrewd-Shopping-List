package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

const authenticatedRule = `@request.auth.id != ""`

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		name := users.Fields.GetByName("name").(*core.TextField)
		name.Required = true
		name.Max = 20
		users.Fields.Add(&core.SelectField{Name: "color", Values: []string{"sky", "sun"}, MaxSelect: 1})

		// Everyone signed in can read both accounts, to show the other person's name.
		users.ListRule = types.Pointer(authenticatedRule)
		users.ViewRule = types.Pointer(authenticatedRule)
		users.CreateRule = nil
		users.UpdateRule = types.Pointer("id = @request.auth.id")
		users.DeleteRule = nil
		users.AuthToken.Duration = 90 * 24 * 60 * 60

		if err := app.Save(users); err != nil {
			return err
		}

		lists := core.NewBaseCollection("lists")
		lists.ListRule = types.Pointer(authenticatedRule)
		lists.ViewRule = types.Pointer(authenticatedRule)
		lists.CreateRule = types.Pointer(authenticatedRule + " && @request.body.created_by = @request.auth.id")
		lists.UpdateRule = types.Pointer(authenticatedRule)
		lists.DeleteRule = types.Pointer(authenticatedRule)
		lists.Fields.Add(
			// Local calendar day as text, so a list never shifts day across time zones.
			&core.TextField{Name: "date", Required: true, Pattern: `^\d{4}-\d{2}-\d{2}$`},
			&core.TextField{Name: "title", Required: true, Max: 40},
			&core.RelationField{Name: "created_by", CollectionId: users.Id, Required: true, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		lists.AddIndex("idx_lists_date", false, "date", "")

		if err := app.Save(lists); err != nil {
			return err
		}

		items := core.NewBaseCollection("items")
		items.ListRule = types.Pointer(authenticatedRule)
		items.ViewRule = types.Pointer(authenticatedRule)
		items.CreateRule = types.Pointer(authenticatedRule + " && @request.body.added_by = @request.auth.id")
		// added_by can only become the requester: an item in the cart that is written again
		// goes back to the items to buy in the name of whoever wrote it.
		items.UpdateRule = types.Pointer(authenticatedRule +
			" && (@request.body.added_by:isset = false || @request.body.added_by = @request.auth.id)" +
			` && (@request.body.checked_by:isset = false || @request.body.checked_by = "" || @request.body.checked_by = @request.auth.id)`)
		items.DeleteRule = types.Pointer(authenticatedRule)
		items.Fields.Add(
			&core.RelationField{Name: "list", CollectionId: lists.Id, Required: true, CascadeDelete: true, MaxSelect: 1},
			&core.TextField{Name: "name", Required: true, Max: 60},
			&core.TextField{Name: "qty", Max: 20},
			&core.BoolField{Name: "checked"},
			&core.DateField{Name: "checked_at"},
			&core.RelationField{Name: "added_by", CollectionId: users.Id, Required: true, MaxSelect: 1},
			&core.RelationField{Name: "checked_by", CollectionId: users.Id, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		items.AddIndex("idx_items_list", false, "list", "")

		if err := app.Save(items); err != nil {
			return err
		}

		// No API rules: only the custom push routes and the items hook touch this collection.
		subscriptions := core.NewBaseCollection("push_subscriptions")
		subscriptions.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, Required: true, CascadeDelete: true, MaxSelect: 1},
			&core.TextField{Name: "endpoint", Required: true},
			&core.TextField{Name: "p256dh", Required: true},
			&core.TextField{Name: "auth", Required: true},
			&core.TextField{Name: "user_agent"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		subscriptions.AddIndex("idx_push_subscriptions_endpoint", true, "endpoint", "")

		return app.Save(subscriptions)
	}, func(app core.App) error {
		for _, collectionName := range []string{"push_subscriptions", "items", "lists"} {
			collection, err := app.FindCollectionByNameOrId(collectionName)
			if err != nil {
				return err
			}
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Restores the defaults of the PocketBase init migration.
		name := users.Fields.GetByName("name").(*core.TextField)
		name.Required = false
		name.Max = 255
		users.Fields.RemoveByName("color")

		ownerRule := "id = @request.auth.id"
		users.ListRule = types.Pointer(ownerRule)
		users.ViewRule = types.Pointer(ownerRule)
		users.CreateRule = types.Pointer("")
		users.UpdateRule = types.Pointer(ownerRule)
		users.DeleteRule = types.Pointer(ownerRule)
		users.AuthToken.Duration = 5 * 24 * 60 * 60

		return app.Save(users)
	})
}
