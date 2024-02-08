package skipper

import (
	"path/filepath"
	"strings"
)

func PathFileBaseName(path string) string {
	str := filepath.Base(path)
	str = strings.TrimSuffix(str, filepath.Ext(str))

	return str
}
