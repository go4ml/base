package fu

import (
	"go-ml.dev/pkg/iokit"
	"path/filepath"
)

func ModelPath(s string) string {
	if filepath.IsAbs(s) {
		return s
	}
	return iokit.CacheFile(filepath.Join("go-ml", "Models", s))
}
