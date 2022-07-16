package templater

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

	targetFiles, err := inv.discoverFiles(targetPath)
	if err != nil {
		return err
	}

	// load all target files the same way as we load the class files
	for _, target := range targetFiles {
		err = target.Load(inv.fs)
		if err != nil {
			return err
		}

		ta := &TargetConfig{}
		target.LoadAs(inv.fs, ta)
		log.Println(ta)

		relativePath := strings.ReplaceAll(target.Path, targetPath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		t, err := NewTarget(target, relativePath)
		if err != nil {
			return fmt.Errorf("unable to create target from file: %s: %w", target.Path, err)
		}
		inv.targetFiles = append(inv.targetFiles, t)

		log.Println("target file:", t)
	}

	// TODO: How to handle if imported classes define the same keys?
	// Maybe just overwrite based on the 'use' order in the target?

	/*
		// Sort data files by path-depth.
		// They are sorted in order to allow for 'depth-based' overwriting. The idea behind that is that
		// you define the more general data in the upper directories (say 'common.yaml') and specify it
		// while going 'down the path'. This allows you to overwrite common data easily.
		sort.SliceStable(inv.files, func(i, j int) bool {
			return strings.Count(inv.files[i].Path, string(os.PathSeparator)) < strings.Count(inv.files[j].Path, string(os.PathSeparator))
		})

		for _, dataFile := range inv.files {
			relativePath := strings.ReplaceAll(dataFile.Path, dataPath, "")
			relativePath = strings.TrimLeft(relativePath, "/")
			class, err := NewClass(dataFile, relativePath)
			if err != nil {
				return fmt.Errorf("unable to create class from data file: %s: %w", dataFile, err)
			}

			log.Println(class)
		}

		// Merge the sorted data files into one Data map
		var data Data
		for i := 0; i < len(inv.files); i++ {
			err := inv.files[i].Load(inv.fs)
			if err != nil {
				return err
			}
			data = mergeData(data, inv.files[i].Data)
		}
		inv.Data = data

	*/

	return nil
}

func (inv *Inventory) TargetExists(name string) bool {
	for _, target := range inv.targetFiles {
		if strings.ToLower(name) == strings.ToLower(target.Name) {
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

func (inv *Inventory) matchesExtension(path string) bool {
	ext := filepath.Ext(path)
	for _, extension := range inv.fileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}
