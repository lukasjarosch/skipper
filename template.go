package skipper

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
)

var customFuncs map[string]any = map[string]any{
	"tfStringArray": func(input []interface{}) string {
		var s []string
		for _, v := range input {
			s = append(s, "\""+fmt.Sprintf(v.(string))+"\"")
		}
		return strings.Join(s, ", ")
	},
}

type Templater struct {
	Files            []*TemplateFile
	templateRootPath string
	outputRootPath   string
	templateFs       afero.Fs
	outputFs         afero.Fs
}

func NewTemplater(fileSystem afero.Fs, templateRootPath, outputRootPath string, userFuncMap map[string]any) (*Templater, error) {
	t := &Templater{
		templateFs: afero.NewBasePathFs(fileSystem, templateRootPath),
		outputFs:   afero.NewBasePathFs(fileSystem, outputRootPath),
	}

	// perpare template functions
	templateFunctions := sprig.TxtFuncMap()

	// merge our own custom functions
	for key, customFunc := range customFuncs {
		templateFunctions[key] = customFunc
	}
	// merge userFuncMap
	if userFuncMap != nil {
		for key, customFunc := range userFuncMap {
			templateFunctions[key] = customFunc
		}
	}

	err := afero.Walk(t.templateFs, "", func(filePath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		file, err := NewTemplateFile(filePath, templateFunctions)
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

// Execute is responsible of parsing and executing the given template, using the passed data context.
// If execution is successful, the template is written to it's desired target location.
// If allowNoValue is true, the template is rendered even if it contains variables which are not defined.
func (t *Templater) Execute(template *TemplateFile, data any, allowNoValue bool) error {
	err := template.Parse(t.templateFs)
	if err != nil {
		return err
	}

	out := new(bytes.Buffer)
	err = template.Execute(out, data)
	if err != nil {
		return err
	}

	// no value detection using scanner in order to give a rough estimate on where the original error is
	if !allowNoValue {
		scanner := bufio.NewScanner(strings.NewReader(out.String()))

		line := 1

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "<no value>") {
				return fmt.Errorf("template '%s' uses variables with undefined value on line %d (line number is based on the rendered output, not the template, search in proximity)", template.Path, line)
			}
			line++
		}
	}

	err = t.writeOutputFile(out.Bytes(), template.Path)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteAll is just a convenience function to execute all templates in `Templater.Files`
func (t *Templater) ExecuteAll(data any, allowNoValue bool) error {
	for _, template := range t.Files {
		err := t.Execute(template, data, allowNoValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeOutputFile ensures that `filePath` exists in the `outputFs` and then writes `data` into it.
func (t *Templater) writeOutputFile(data []byte, filePath string) error {
	err := t.outputFs.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	err = afero.WriteFile(t.outputFs, filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
