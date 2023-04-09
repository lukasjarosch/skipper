package skipper

import (
	"errors"
	"strings"
)

var (
	ErrEmptyDataPath error = errors.New("empty DataPath")
)

const DataPathSeparator = "."

type DataPath []string

func DataPathFromString(path string) (DataPath, error) {
	if path == "" {
		return nil, ErrEmptyDataPath
	}

	return strings.Split(path, DataPathSeparator), nil
}

func (dp DataPath) String() string {
	return strings.Join(dp, DataPathSeparator)
}

type Data map[string]interface{}
