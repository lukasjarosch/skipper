package skipper

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// DiscoverYamlFiles iterates over a given rootPath recursively, filters out all files with the appropriate file fileExtensions
// and finally creates a YamlFile slice which is then returned.
func DiscoverYamlFiles(fileSystem afero.Fs, rootPath string) ([]*YamlFile, error) {
	exists, err := afero.Exists(fileSystem, rootPath)
	if err != nil {
		return nil, fmt.Errorf("check if path exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("file path does not exist: %s", rootPath)
	}

	matchesExtension := func(path string) bool {
		ext := filepath.Ext(path)
		for _, extension := range yamlFileExtensions {
			if extension == ext {
				return true
			}
		}
		return false
	}

	var files []*YamlFile
	err = afero.Walk(fileSystem, rootPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if matchesExtension(path) {
			file, err := NewYamlFile(path)
			if err != nil {
				return err
			}
			files = append(files, file)
		}
		return nil
	})

	return files, nil
}

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

// CopyFileFsToFs will copy a file from the given sourceFs and sourcePath to the targetFs and targetPath
func CopyFileFsToFs(sourceFs afero.Fs, targetFs afero.Fs, sourcePath, targetPath string) error {
	sourceExists, err := afero.Exists(sourceFs, sourcePath)
	if err != nil {
		return fmt.Errorf("CopyFileFsToFs failed to check source file: %w", err)
	}
	if !sourceExists {
		return fmt.Errorf("CopyFileFsToFs source path: %s: %w", sourcePath, os.ErrNotExist)
	}

	err = targetFs.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		return fmt.Errorf("CopyFileFsToFs failed to create path: %w", err)
	}

	sourceData, err := afero.ReadFile(sourceFs, sourcePath)
	if err != nil {
		return fmt.Errorf("CopyFileFsToFs failed to read source file: %w", err)
	}

	targetFile, err := targetFs.Create(targetPath)
	if err != nil {
		return fmt.Errorf("CopyFileFsToFs failed to create target file: %w", err)
	}
	defer targetFile.Close()

	bytesWritten, err := targetFile.Write(sourceData)
	if err != nil {
		return fmt.Errorf("CopyFileFsToFs failed to write target file: %w", err)
	}
	if bytesWritten != len(sourceData) {
		return fmt.Errorf("CopyFileFsToFs did not write all source file bytes into the target file, retry")
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
