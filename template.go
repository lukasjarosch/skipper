package skipper

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"mime"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

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

	"addDate": func(years, months, days string) string {
		y, err := strconv.Atoi(years)
		if err != nil {
			return err.Error()
		}
		m, err := strconv.Atoi(months)
		if err != nil {
			return err.Error()
		}
		d, err := strconv.Atoi(days)
		if err != nil {
			return err.Error()
		}
		return time.Now().AddDate(y, m, d).Format(time.RFC3339)
	},

	"context": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("uneven amount of values")
		}
		context := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("map key must be a string")
			}
			context[key] = values[i+1]
		}

		return context, nil
	},
}

type Templater struct {
	Files            []*File
	partialTemplates []*File
	IgnoreRegex      []*regexp.Regexp
	templateRootPath string
	outputRootPath   string
	templateFs       afero.Fs
	outputFs         afero.Fs
	templateFuncs    template.FuncMap
}

func NewTemplater(fileSystem afero.Fs, templateRootPath, outputRootPath string, userFuncMap map[string]any, ignoreRegex []string) (*Templater, error) {
	t := &Templater{
		templateFs:    afero.NewBasePathFs(fileSystem, templateRootPath),
		outputFs:      afero.NewBasePathFs(fileSystem, outputRootPath),
		templateFuncs: sprig.TxtFuncMap(),
	}

	// merge our own custom functions
	for key, customFunc := range customFuncs {
		t.templateFuncs[key] = customFunc
	}
	// merge userFuncMap
	if userFuncMap != nil {
		for key, customFunc := range userFuncMap {
			t.templateFuncs[key] = customFunc
		}
	}

	exists, err := afero.DirExists(fileSystem, templateRootPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("templateRootPath does not exist: %s", templateRootPath)
	}

	// Save the regexp
	for _, v := range ignoreRegex {
		r := regexp.MustCompile(v)
		t.IgnoreRegex = append(t.IgnoreRegex, r)
	}

	// discover all files in the templateRootPath
	err = afero.Walk(t.templateFs, "", func(filePath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		// Not all files are actually templates, there may very well exist other files (such as images)
		// which we do not want to register as templates as they would not parse.
		ignoredMimeTypes := []*regexp.Regexp{
			regexp.MustCompile(`image/.*`),
			regexp.MustCompile(`video/.*`),
		}
		mimeType := mime.TypeByExtension(filepath.Ext(info.Name()))

		for _, ignoredMimeRegex := range ignoredMimeTypes {
			if !ignoredMimeRegex.MatchString(mimeType) {
				continue
			}
			return nil // skip this file
		}

		// Load and register template file
		file, err := NewFile(filePath)
		if err != nil {
			return err
		}
		err = file.Load(t.templateFs)
		if err != nil {
			return err
		}
		t.Files = append(t.Files, file)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking over template path: %w", err)
	}

	t.DiscoverPartials()

	return t, nil
}

// DiscoverPartials will iterate over all registered files and check each of them
// whether an additional template is defined (e.g. using the 'define' directive).
// If so, the file is added to the list of partial templates which is made available
// during template execution.
// This ensures that every template can access partial templates.
// Subsequent calls to this method will reset the 'partialTemplates' field.
func (t *Templater) DiscoverPartials() {
	t.partialTemplates = []*File{}

	for _, tplFile := range t.Files {
		if t.isPathIgnored(tplFile.Path) {
			continue
		}

		// There exists at least one template named 'test'.
		// If there is more than one, the file contains a partial template definition.
		test := template.New(tplFile.Path).Funcs(t.templateFuncs)
		test.Parse(string(tplFile.Bytes))
		if len(test.Templates()) == 1 {
			continue
		}
		t.partialTemplates = append(t.partialTemplates, tplFile)
	}
}

// DefaultTemplateContext returns the default template context with an 'Inventory' field where the Data is located.
// Additionally it adds the 'TargetName' field for convenience.
func DefaultTemplateContext(data Data, targetName string) any {
	return struct {
		Inventory  any
		TargetName string
	}{
		Inventory:  data,
		TargetName: targetName,
	}
}

// Execute is responsible of parsing and executing the given template, using the passed data context.
// If execution is successful, the template is written to it's desired target location.
// If allowNoValue is true, the template is rendered even if it contains variables which are not defined.
func (t *Templater) Execute(template *File, data any, allowNoValue bool, renameConfig []RenameConfig) error {
	// if a renameConfig exists, rename possible files accordingly
	if renameConfig != nil {
		for _, rename := range renameConfig {
			if strings.EqualFold(template.Path, rename.InputPath) {
				return t.execute(template, data, rename.Filename, allowNoValue)
			}
		}
	}
	return t.execute(template, data, template.Path, allowNoValue)
}

// execute is the main rendering function for templates
func (t *Templater) execute(tplFile *File, data any, targetPath string, allowNoValue bool) error {
	// if the template matches any IgnoreRegex, just copy the file to the targetPath
	// without rendering it as template
	for _, v := range t.IgnoreRegex {
		if v.MatchString(tplFile.Path) {
			err := CopyFileFsToFs(t.templateFs, t.outputFs, tplFile.Path, targetPath)
			if err != nil {
				return fmt.Errorf("could not copy file %s: %w", tplFile.Path, err)
			}
			return nil
		}
	}

	// create new target template with the attached functions
	tpl := template.New(tplFile.Path).Funcs(t.templateFuncs)

	// Add every discovered partial template in case its needed.
	for _, partialTemplate := range t.partialTemplates {
		if t.isPathIgnored(partialTemplate.Path) {
			continue
		}

		var err error
		tpl, err = tpl.Parse(string(partialTemplate.Bytes))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", partialTemplate.Path, err)
		}
	}

	// now, finally parse the template we're actually aiming to render
	var err error
	tpl, err = tpl.Parse(string(tplFile.Bytes))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tplFile.Path, err)
	}
	out := new(bytes.Buffer)
	err = tpl.Execute(out, data)
	if err != nil {
		return err
	}

	// no value detection using scanner in order to give a rough estimate on where the original error is
	if !allowNoValue {
		scanner := bufio.NewScanner(strings.NewReader(out.String()))

		line := 1

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "<no value>") {
				return fmt.Errorf("template '%s' uses variables with undefined value on line %d (line number is based on the rendered output and might not be accurate)", tplFile.Path, line)
			}
			line++
		}
	}

	err = WriteFile(t.outputFs, targetPath, out.Bytes(), tplFile.Mode)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteComponents will only execute the templates as they are defined in the given components.
func (t *Templater) ExecuteComponents(data any, components []ComponentConfig, allowNoValue bool) error {
	if len(components) == 0 {
		return fmt.Errorf("no components to render")
	}

	for _, component := range components {
		for _, input := range component.InputPaths {
			file := t.getTemplateByPath(input)

			if file == nil {
				continue
			}

			outputFileName := filepath.Base(file.Path)

			// if the input path has a rename config, we need to change the outputFileName accordingly
			for _, rename := range component.Renames {
				if strings.EqualFold(rename.InputPath, input) {
					outputFileName = rename.Filename
				}
			}

			err := t.execute(file, data, filepath.Join(component.OutputPath, outputFileName), allowNoValue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ExecuteAll is just a convenience function to execute all templates in `Templater.Files`
func (t *Templater) ExecuteAll(data any, allowNoValue bool, renameConfig []RenameConfig) error {
	for _, template := range t.Files {
		err := t.Execute(template, data, allowNoValue, renameConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Templater) getTemplateByPath(path string) *File {
	for _, file := range t.Files {
		if file.Path == path {
			return file
		}
	}
	return nil
}

// isPathIgnored provides a quick way to check whether a given path should be
// ignored based on the 'IgnoreRegex' field.
func (t *Templater) isPathIgnored(filePath string) bool {
	for _, v := range t.IgnoreRegex {
		if v.MatchString(filePath) {
			return true
		}
	}
	return false
}
