//go:build embed

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:pb_public
var embeddedPublic embed.FS

// publicFiles is the Astro build compiled into the binary, for production (go build -tags embed).
func publicFiles() fs.FS {
	files, err := fs.Sub(embeddedPublic, "pb_public")
	if err != nil {
		panic(err)
	}
	return files
}
