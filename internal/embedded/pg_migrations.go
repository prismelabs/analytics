package embedded

import "embed"

//go:embed pg_migrations
var PgMigrations embed.FS
