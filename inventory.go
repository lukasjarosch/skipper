package skipper

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Inventory is the collection of data files.
type Inventory struct {
	fs             afero.Fs
	fileExtensions []string
	classFiles     []*Class
	targetFiles    []*Target
	Data           Data
}

// NewInventory creates a new Inventory with the given afero.Fs.
// At least one extension must be provided, otherwise an error is returned.
func NewInventory(fs afero.Fs) (*Inventory, error) {
	if fs == nil {
		return nil, fmt.Errorf("fs cannot be nil")
	}

	inv := &Inventory{
		fs:             fs,
		fileExtensions: []string{".yml", ".yaml"},
	}

	return inv, nil
}

func (inv *Inventory) Load(classPath, targetPath string) error {

	err := inv.loadClassFiles(classPath)
	if err != nil {
		return fmt.Errorf("unable to load class files: %w", err)
	}

	err = inv.loadTargetFiles(targetPath)
	if err != nil {
		return fmt.Errorf("unable to load target files: %w", err)
	}

	// check for all targets whether they use classes which actually exist
	for _, target := range inv.targetFiles {
		for _, class := range target.UsedClasses {
			if !inv.ClassExists(class) {
				return fmt.Errorf("target '%s' uses class '%s' which does not exist", target.Name, class)
			}
		}
	}

	// TODO: How to handle if imported classes define the same keys?
	// Maybe just overwrite based on the 'use' order in the target?

	return nil
}

func (inv *Inventory) Target(name string) (*Target, error) {
	if !inv.TargetExists(name) {
		return nil, fmt.Errorf("target '%s' does not exist", name)
	}

	return inv.getTarget(name), nil
}

func (inv *Inventory) TargetExists(name string) bool {
	if inv.getTarget(name) == nil {
		return false
	}
	return true
}

func (inv *Inventory) getTarget(name string) *Target {
	for _, target := range inv.targetFiles {
		if strings.ToLower(name) == strings.ToLower(target.Name) {
			return target
		}
	}
	return nil
}

func (inv *Inventory) ClassExists(name string) bool {
	for _, class := range inv.classFiles {
		if class.Name == name {
			return true
		}
	}
	return false
}

func (inv *Inventory) discoverFiles(rootPath string) ([]*YamlFile, error) {
	exists, err := afero.Exists(inv.fs, rootPath)
	if err != nil {
		return nil, fmt.Errorf("check if path exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("file path does not exist: %s", rootPath)
	}

	var files []*YamlFile
	err = afero.Walk(inv.fs, rootPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if inv.matchesExtension(path) {
			file, err := NewFile(path)
			if err != nil {
				return err
			}
			files = append(files, file)
		}
		return nil
	})
	return files, err
}

func (inv *Inventory) loadClassFiles(classPath string) error {
	classFiles, err := inv.discoverFiles(classPath)
	if err != nil {
		return err
	}

	// load all class files, replacing the inventory-relative path with dot-separated style
	for _, class := range classFiles {
		err = class.Load(inv.fs)
		if err != nil {
			return err
		}

		relativePath := strings.ReplaceAll(class.Path, classPath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		c, err := NewClass(class, relativePath)
		if err != nil {
			return fmt.Errorf("unable to create class from file: %s: %w", class.Path, err)
		}
		inv.classFiles = append(inv.classFiles, c)
	}
	return nil
}

func (inv *Inventory) loadTargetFiles(targetPath string) error {
	targetFiles, err := inv.discoverFiles(targetPath)
	if err != nil {
		return err
	}

	for _, target := range targetFiles {
		err = target.Load(inv.fs)
		if err != nil {
			return err
		}

		relativePath := strings.ReplaceAll(target.Path, targetPath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		t, err := NewTarget(target, relativePath)
		if err != nil {
			return fmt.Errorf("%s: %w", target.Path, err)
		}
		inv.targetFiles = append(inv.targetFiles, t)

		log.Println(t.UsedClasses)
	}

	return nil
}

func (inv *Inventory) matchesExtension(path string) bool {
	ext := filepath.Ext(path)
	for _, extension := range inv.fileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}
