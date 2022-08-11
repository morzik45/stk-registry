//go:build !prod
// +build !prod

package frontimport

import (
	"io/fs"
	"os"
)

func GetFrontendAssets() fs.FS {
	return os.DirFS("dist")
}
