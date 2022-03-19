package fu

import (
	"go4ml.xyz/iokit"
	"path/filepath"
)

func ModelPath(s string) string {
	if filepath.IsAbs(s) {
		return s
	}
	return iokit.CacheFile(filepath.Join("go-ml", "Models", s))
}
