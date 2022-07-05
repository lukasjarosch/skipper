package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

// Inventory is the collection of data files.
type Inventory struct {
	fs             afero.Fs
	fileExtensions []string
	files          []*YamlFile
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

func (inv *Inventory) Load(dataPath string) error {
	if err := inv.discoverDataFiles(dataPath); err != nil {
		return err
	}

	// Sort data files by path-depth.
	// They are sorted in order to allow for 'depth-based' overwriting. The idea behind that is that
	// you define the more general data in the upper directories (say 'common.yaml') and specify it
	// while going 'down the path'. This allows you to overwrite common data easily.
	sort.SliceStable(inv.files, func(i, j int) bool {
		return strings.Count(inv.files[i].Path, string(os.PathSeparator)) < strings.Count(inv.files[j].Path, string(os.PathSeparator))
	})

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

	return nil
}

func (inv *Inventory) discoverDataFiles(rootPath string) error {
	exists, err := afero.Exists(inv.fs, rootPath)
	if err != nil {
		return fmt.Errorf("failed to check data path: %w", err)
	}
	if !exists {
		return fmt.Errorf("data path does not exist: %s", rootPath)
	}

	return afero.Walk(inv.fs, rootPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if inv.matchesExtension(path) {
			file, err := NewFile(path)
			if err != nil {
				return err
			}
			inv.files = append(inv.files, file)
		}
		return nil
	})
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
