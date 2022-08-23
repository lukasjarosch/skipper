package skipper

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// file is just an arbitrary description of a path and the data of the file to which Path points to.
// Note that the used filesystem is not relevant, only at the time of loading a file.
type file struct {
	Path  string
	Bytes []byte
}

func newFile(path string) (*file, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	return &file{Path: path}, nil
}

// Load will attempt to read the file from the given filesystem implementation.
// The loaded data is stored in `file.Bytes`
func (f *file) Load(fs afero.Fs) (err error) {
	f.Bytes, err = afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load %s: %w", f.Path, err)
	}
	return nil
}

// YamlFile is what is used for all inventory-relevant files (classes and targets).
type YamlFile struct {
	file
	Data Data
}

// NewFile returns a newly initialized `YamlFile`.
func NewFile(path string) (*YamlFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &YamlFile{
		file: *f,
	}, nil
}

// CreateNewFile can be used to manually create a file inside the given filesystem.
// This is useful for dynamically creating classes or targets.
//
// The given path is attempted to be created and a file written.
func CreateNewFile(fs afero.Fs, path string, data []byte) (*YamlFile, error) {
	err := fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	err = afero.WriteFile(fs, path, data, 0644)
	if err != nil {
		return nil, err
	}
	return NewFile(path)
}

// Load will first load the underlying raw file-data and then attempt to `yaml.Unmarshal` it into `Data`
// The resulting Data is stored in `YamlFile.Data`.
func (f *YamlFile) Load(fs afero.Fs) error {
	err := f.file.Load(fs)
	if err != nil {
		return err
	}

	var d Data
	if err := yaml.Unmarshal(f.Bytes, &d); err != nil {
		return err
	}
	f.Data = d
	return nil
}

// TemplateFile represents a file which is used as Template.
type TemplateFile struct {
	file
	tpl *template.Template
}

// NewTemplateFile creates a new TemplateFile at the given path.
// User-defined template funcs can be added to have them available in the templates.
// By default, the well-known sprig functions are always added (see: https://github.com/Masterminds/sprig).
func NewTemplateFile(path string, funcs map[string]any) (*TemplateFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &TemplateFile{
		file: *f,
		tpl:  template.New(path).Funcs(sprig.TxtFuncMap()).Funcs(funcs),
	}, nil
}

// Parse will attempt to Load and parse the template from the given filesystem.
func (tmpl *TemplateFile) Parse(fs afero.Fs) (err error) {
	err = tmpl.file.Load(fs)
	if err != nil {
		return err
	}

	tmpl.tpl, err = tmpl.tpl.Parse(string(tmpl.Bytes))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmpl.Path, err)
	}

	return nil
}

// Execute renders the template file and writes the output into the passed `io.Writer`
// The passed contexData is what will be available inside the template.
func (tmpl *TemplateFile) Execute(out io.Writer, contextData any) (err error) {
	return tmpl.tpl.Execute(out, contextData)
}
