package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

type Config struct {
	SourcePaths []string `yaml:"sourcePaths"`
	TargetPath  string   `yaml:"targetPath"`
	Overwrite   bool     `yaml:"overwrite"`
}

func NewConfig(raw map[string]interface{}) (Config, error) {
	yaml := codec.NewYamlCodec()
	b, err := yaml.Marshal(raw)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.UnmarshalTarget(b, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

var ConfigPath = data.NewPath("config.plugins.output.copy")

// Plugin symbol which is loaded by skipper.
// If this does not exist, the plugin cannot be used.
var Plugin = NewCopy()

type Copy struct {
	config Config
}

func NewCopy() skipper.PluginConstructor {
	return func() skipper.Plugin {
		return &Copy{}
	}
}

// Configure checks that all configured source-paths exist and are readable.
func (copy *Copy) Configure() error {
	// source-paths must be readable
	for _, sourcePath := range copy.config.SourcePaths {
		_, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("cannot open source-path: %w", err)
		}
	}

	return nil
}

func (copy *Copy) Run() error {
	for _, source := range copy.config.SourcePaths {
		err := copyST(source, copy.config.TargetPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConfigPointer returns a pointer to [Config] which is going to be used by skipper
// to automatically load the config into said pointer, nice.
func (copy *Copy) ConfigPointer() interface{} {
	return &copy.config
}

func (copy *Copy) Name() string {
	return "copy"
}

func (copy *Copy) Type() skipper.PluginType {
	return skipper.OutputPlugin
}

// Copy recursively copies the contents of sourcePath to targetPath.
func copyST(sourcePath, targetPath string) error {
	// Get the file information of the source
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	// If the source is a directory, call the directory copy function
	if sourceInfo.IsDir() {
		return copyDir(sourcePath, targetPath)
	}

	// If the source is a file, call the file copy function
	return copyFile(sourcePath, targetPath)
}

// copyDir copies the contents of a directory from sourcePath to targetPath.
func copyDir(sourcePath, targetPath string) error {
	// Create the target directory with the same permissions as source directory
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(targetPath, sourceInfo.Mode())
	if err != nil {
		return err
	}

	// Read the contents of the source directory
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return err
	}

	// Recursively copy each entry in the source directory
	for _, entry := range entries {
		sourceEntryPath := filepath.Join(sourcePath, entry.Name())
		targetEntryPath := filepath.Join(targetPath, entry.Name())
		err = copyST(sourceEntryPath, targetEntryPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a single file from sourcePath to targetPath.
func copyFile(sourcePath, targetPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
