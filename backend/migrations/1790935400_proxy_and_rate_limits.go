package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// In production PocketBase listens on localhost behind Caddy, which appends the client address
// to X-Forwarded-For: trusting its rightmost value gives the rate limits the real client IP.
// Without the proxy, as in development, the header is missing and the remote address is used.
func init() {
	m.Register(func(app core.App) error {
		settings := app.Settings()
		settings.TrustedProxy.Headers = []string{"X-Forwarded-For"}
		// The default rules, with 2 sign-in attempts every 3 seconds per address.
		settings.RateLimits.Enabled = true

		return app.Save(settings)
	}, func(app core.App) error {
		settings := app.Settings()
		settings.TrustedProxy.Headers = nil
		settings.RateLimits.Enabled = false

		return app.Save(settings)
	})
}
