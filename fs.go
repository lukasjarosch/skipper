package skipper

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// CopyFile will copy the given sourcePath to the targetPath inside the passed afero.Fs.
func CopyFile(fs afero.Fs, sourcePath, targetPath string) error {
	sourceExists, err := afero.Exists(fs, sourcePath)
	if err != nil {
		return fmt.Errorf("CopyFile failed to check source file: %w", err)
	}
	if !sourceExists {
		return fmt.Errorf("CopyFile source path: %s: %w", sourcePath, os.ErrNotExist)
	}

	err = fs.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		return fmt.Errorf("CopyFile failed to create path: %w", err)
	}

	sourceData, err := afero.ReadFile(fs, sourcePath)
	if err != nil {
		return fmt.Errorf("CopyFile failed to read source file: %w", err)
	}

	targetFile, err := fs.Create(targetPath)
	if err != nil {
		return fmt.Errorf("CopyFile failed to create target file: %w", err)
	}
	defer targetFile.Close()

	bytesWritten, err := targetFile.Write(sourceData)
	if err != nil {
		return fmt.Errorf("CopyFile failed to write target file: %w", err)
	}
	if bytesWritten != len(sourceData) {
		return fmt.Errorf("CopyFile did not write all source file bytes into the target file, retry")
	}

	return nil
}

// WriteFile ensures that `targetPath` exists in the `fs` and then writes `data` into it.
func WriteFile(fs afero.Fs, targetPath string, data []byte, mode fs.FileMode) error {
	err := fs.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		return err
	}

	err = afero.WriteFile(fs, targetPath, data, mode)
	if err != nil {
		return err
	}

	return nil
}
