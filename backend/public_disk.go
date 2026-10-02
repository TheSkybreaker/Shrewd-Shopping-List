//go:build !embed

package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/tools/osutils"
)

// publicFiles is the Astro build on disk, for development: next to the binary, or in the working
// directory under `go run`.
func publicFiles() fs.FS {
	if osutils.IsProbablyGoRun() {
		return os.DirFS("./pb_public")
	}
	return os.DirFS(filepath.Join(os.Args[0], "../pb_public"))
}
