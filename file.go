package skipper

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// File is just an arbitrary description of a path and the data of the file to which Path points to.
// Note that the used filesystem is not relevant, only at the time of loading a file.
type File struct {
	Path  string
	Bytes []byte
}

func newFile(path string) (*File, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	return &File{Path: path}, nil
}

// Exists returns true if the file exists in the given filesystem, false otherwise.
func (f *File) Exists(fs afero.Fs) bool {
	exists, err := afero.Exists(fs, f.Path)
	if err != nil {
		return false
	}
	return exists
}

// Load will attempt to read the file from the given filesystem implementation.
// The loaded data is stored in `File.Bytes`
func (f *File) Load(fs afero.Fs) (err error) {
	if fs == nil {
		return fmt.Errorf("fs cannot be nil")
	}

	f.Bytes, err = afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load %s: %w", f.Path, err)
	}
	return nil
}

// YamlFile is what is used for all inventory-relevant files (classes and targets).
type YamlFile struct {
	File
	Data Data
}

// NewYamlFile returns a newly initialized `YamlFile`.
func NewYamlFile(path string) (*YamlFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &YamlFile{
		File: *f,
	}, nil
}

// CreateNewYamlFile can be used to manually create a file inside the given filesystem.
// This is useful for dynamically creating classes or targets.
//
// The given path is attempted to be created and a file written.
func CreateNewYamlFile(fs afero.Fs, path string, data []byte) (*YamlFile, error) {
	if fs == nil {
		return nil, fmt.Errorf("fs cannot be nil")
	}

	err := fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	err = afero.WriteFile(fs, path, data, 0644)
	if err != nil {
		return nil, err
	}
	return NewYamlFile(path)
}

// Load will first load the underlying raw file-data and then attempt to `yaml.Unmarshal` it into `Data`
// The resulting Data is stored in `YamlFile.Data`.
func (f *YamlFile) Load(fs afero.Fs) error {
	err := f.File.Load(fs)
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

// UnmarshalPath can be used to unmarshall only a sub-map of the Data inside [YamlFile].
// The function errors if the file has not been loaded.
func (f *YamlFile) UnmarshalPath(target interface{}, path ...interface{}) error {
	if f.Data == nil {
		return fmt.Errorf("yaml file not loaded, no data exists")
	}
	data, err := f.Data.GetPath(path...)
	if err != nil {
		return err
	}

	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, target)
}

// TemplateFile represents a file which is used as Template.
type TemplateFile struct {
	File
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

	// If funcs are passed, we need to ensure that all values are functions otherwise text/template will panic
	// There is an additional limitation: The function must return either a single (string) or two (string, error) values.
	// We are not going to validate this as one just needs to read the docs.
	for k, v := range funcs {
		typ := reflect.TypeOf(v).Kind()
		if typ != reflect.Func {
			return nil, fmt.Errorf("funcs[%s] is not a function", k)
		}
	}

	return &TemplateFile{
		File: *f,
		tpl:  template.New(path).Funcs(sprig.TxtFuncMap()).Funcs(funcs),
	}, nil
}

// Parse will attempt to Load and parse the template from the given filesystem.
func (tmpl *TemplateFile) Parse(fs afero.Fs) (err error) {
	err = tmpl.File.Load(fs)
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

type SecretFile struct {
	*YamlFile
	RelativePath string
}

type SecretFileList []*SecretFile

func NewSecretFile(file *YamlFile, relativeSecretPath string) (*SecretFile, error) {
	return &SecretFile{
		YamlFile:     file,
		RelativePath: relativeSecretPath,
	}, nil
}

func (sfl SecretFileList) GetSecretFile(path string) *SecretFile {
	for _, secretFile := range sfl {
		if strings.EqualFold(secretFile.RelativePath, path) {
			return secretFile
		}
	}
	return nil
}
