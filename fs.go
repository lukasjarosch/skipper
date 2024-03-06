package skipper

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PathFileBaseName returns the basename of the file from the given path.
// The file extension is removed.
func PathFileBaseName(path string) string {
	str := filepath.Base(path)
	str = strings.TrimSuffix(str, filepath.Ext(str))

	return str
}

// DiscoverFiles walks down the given rootPath and emits a list of files which match the pathSelector.
func DiscoverFiles(rootPath string, pathSelector *regexp.Regexp) ([]string, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("rootPath is empty")
	}
	if pathSelector == nil {
		return nil, fmt.Errorf("selectorRegex is nil")
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("rootPath is not a directory")
	}

	var files []string
	err = filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if pathSelector.MatchString(path) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// CommonPathPrefix determines the longest common prefix of the given paths.
func CommonPathPrefix(paths []string) string {
	prefix := paths[0]
	for _, path := range paths {
		for !strings.HasPrefix(path, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

// StripCommonPathPrefix will detect the longest matching prefix
// among the elements of the paths list and remove it.
func StripCommonPathPrefix(paths []string) []string {
	prefix := CommonPathPrefix(paths)

	strippedPaths := make([]string, len(paths))
	for i, path := range paths {
		strippedPaths[i] = strings.TrimPrefix(path, prefix)
	}

	return strippedPaths
}
