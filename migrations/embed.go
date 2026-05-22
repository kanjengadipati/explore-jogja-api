package migrations

import "embed"

// Files embeds SQL migrations so they are available in serverless and packaged deployments.
//
//go:embed *.sql mysql/*.sql
var Files embed.FS
