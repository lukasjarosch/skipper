package skipper

import (
	"fmt"
	"io"
	"text/template"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

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

func (f *file) Load(fs afero.Fs) (err error) {
	f.Bytes, err = afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load %s: %w", f.Path, err)
	}
	return nil
}

type YamlFile struct {
	file
	Data Data
}

// TODO: change name to NewYamlFile
func NewFile(path string) (*YamlFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &YamlFile{
		file: *f,
	}, nil
}

func CreateNewFile(fs afero.Fs, path string, data []byte) (*YamlFile, error) {
	err := afero.WriteFile(fs, path, data, 0644)
	if err != nil {
		return nil, err
	}
	return NewFile(path)
}

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

type TemplateFile struct {
	file
	tpl *template.Template
}

func NewTemplateFile(path string, funcs map[string]any) (*TemplateFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	tpl := template.New(path).Funcs(funcs)

	return &TemplateFile{
		file: *f,
		tpl:  tpl,
	}, nil
}

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

func (tmpl *TemplateFile) Execute(out io.Writer, data any) (err error) {
	return tmpl.tpl.Execute(out, data)
}
