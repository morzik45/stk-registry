//go:build prod
// +build prod

package frontimport

import (
	"embed"
	"io/fs"
)

//go:embed static
var embedFrontend embed.FS

func GetFrontendAssets() fs.FS {
	f, err := fs.Sub(embedFrontend, "static")
	if err != nil {
		panic(err)
	}

	return f
}
