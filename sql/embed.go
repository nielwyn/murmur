// Package migrations embeds the goose migration files so the server can apply
// its own schema at startup, with no goose binary on the deploy host.
package migrations

import "embed"

//go:embed schema/*.sql
var FS embed.FS
