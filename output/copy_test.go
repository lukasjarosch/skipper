package output_test

import (
	"os"
	"path"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/output"
)

func TestCopy_Run(t *testing.T) {
	fsRoot := "testdata/output/fs"

	createFile := func(t *testing.T, path string, content string) {
		spew.Dump(path)
		err := os.WriteFile(path, []byte(content), 0600)
		assert.NoError(t, err)
	}

	t.Run("source file exists, target does not exist", func(t *testing.T) {
		sourcePath := path.Join(fsRoot, "source")
		createFile(t, sourcePath, "hello")

		// targetPath has no trailing slash!
		targetPath := path.Join(fsRoot, "target")
		(&output.Copy{}).Copy(sourcePath, targetPath)
	})
}
