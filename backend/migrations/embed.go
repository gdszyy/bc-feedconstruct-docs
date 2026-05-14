// Package migrations exposes the SQL migration files as an embedded fs.FS
// so cmd/bffd and cmd/migrate can apply them without depending on the
// runtime working directory or copying files into the container image.
package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var sqlFS embed.FS

// FS returns the embedded migration directory.
func FS() fs.FS { return sqlFS }
