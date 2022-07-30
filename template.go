package skipper

import (
	"fmt"
	"io/fs"

	"github.com/spf13/afero"
)

type Templater struct {
	Files            []*TemplateFile
	templateRootPath string
	fs               afero.Fs
}

func NewTemplater(fileSystem afero.Fs, templateRootPath string) (*Templater, error) {
	t := &Templater{
		fs: afero.NewBasePathFs(fileSystem, templateRootPath),
	}

	err := afero.Walk(t.fs, "", func(filePath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := NewTemplateFile(filePath)
		if err != nil {
			return err
		}
		t.Files = append(t.Files, file)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking over template path: %w", err)
	}

	return t, nil
}

func (t *Templater) Execute() error {
	return nil
}
