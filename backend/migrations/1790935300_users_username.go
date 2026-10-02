package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// The accounts sign in with a username instead of an email, which they no longer need.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Lowercase only: the login page lowercases what the phone keyboard capitalizes.
		users.Fields.Add(&core.TextField{Name: "username", Required: true, Min: 3, Max: 20, Pattern: `^[a-z0-9]+$`})
		// Partial like the email index, so accounts created before this migration can save
		// while their username is still empty.
		users.AddIndex("idx_users_username", true, "username", "username != ''")
		users.PasswordAuth.IdentityFields = []string{"username"}
		users.Fields.GetByName("email").(*core.EmailField).Required = false

		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.PasswordAuth.IdentityFields = []string{"email"}
		users.Fields.GetByName("email").(*core.EmailField).Required = true
		users.RemoveIndex("idx_users_username")
		users.Fields.RemoveByName("username")

		return app.Save(users)
	})
}
