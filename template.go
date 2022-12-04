package skipper

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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
}

type ErrUndefinedTemplateVariable struct {
	error
	DownstreamError    error
	VariableIdentifier string
	TemplateFile       string
	Line               int
}

func (err ErrUndefinedTemplateVariable) Error() string {
	return fmt.Sprintf("undefined template variable '%s' in template '%s' on line %d: %s", err.VariableIdentifier, err.TemplateFile, err.Line, err.DownstreamError)
}

var (
	// matches template variables `{{ ... }}`
	templateVariableRegex = regexp.MustCompile(`(\{\{)(.*)(\}\})`)
	// matches template data identifiers within variables `.Something.else.foo`
	templateVariableIdentifierRegex = regexp.MustCompile(`\.[\.|\w]+\w`)
)

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

	exists, err := afero.DirExists(fileSystem, templateRootPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("templateRootPath does not exist: %s", templateRootPath)
	}

	err = afero.Walk(t.templateFs, "", func(filePath string, info fs.FileInfo, err error) error {
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
func (t *Templater) Execute(template *TemplateFile, data any, allowNoValue bool, renameConfig []RenameConfig) error {
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

func (t *Templater) checkUndefinedTemplateVariables(template *TemplateFile, templateContextData any) error {
	scanner := bufio.NewScanner(bytes.NewBuffer(template.Bytes))
	line := 1
	for scanner.Scan() {
		variableIdentifiers := t.findTemplateVariableIdentifiers(scanner.Bytes())

		if len(variableIdentifiers) == 0 {
			line++
			continue
		}

		// at this point we have variable identifiers like `.Inventory.foo.bar`
		// the first element in the identifier is always the name of a field inside the given [data]
		// we need to figure out, whether this field exists in the [data].
		{
			for _, variableIdentifier := range variableIdentifiers {

				getDataFieldByName := func(name string) reflect.Value {
					tmp := reflect.ValueOf(templateContextData)
					index := 0
					for i := 0; i < tmp.Type().NumField(); i++ {
						fieldName := tmp.Type().Field(i).Name
						if strings.EqualFold(fieldName, name) {
							index = i
							break
						}
					}
					return tmp.FieldByIndex([]int{index})
				}

				identifierToPath := func(identifierElements []string) (path []interface{}) {
					for _, el := range identifierElements {
						path = append(path, el)
					}
					return path
				}

				// value is now most likely (in case of skipper) a map[string]interface{} which we can address
				identifierElements := strings.Split(string(variableIdentifier), ".")
				identifierElements = append(identifierElements[:0], identifierElements[1:]...) // remove first element, it will be empty
				firstIdentifierElement := strings.TrimLeft(strings.TrimSpace(string(identifierElements[0])), ".")
				value := getDataFieldByName(firstIdentifierElement)

				// it is very likely that the value is of type `map[string]interface{}` because that's what we use in skipper.Data
				// hence if there are > 1 identifier elements, we'll assume this case
				if len(identifierElements) > 1 {
					dataKeys := append(identifierElements[:0], identifierElements[1:]...) // data keys begin after the 'firstIdentifierElement'
					if valueData, ok := value.Interface().(Data); ok {
						_, err := valueData.GetPath(identifierToPath(dataKeys)...)
						if err != nil {
							return ErrUndefinedTemplateVariable{
								DownstreamError:    err,
								VariableIdentifier: variableIdentifier,
								TemplateFile:       template.Path,
								Line:               line,
							}
						}
					}
				} else {
					if len(value.String()) == 0 {
						return ErrUndefinedTemplateVariable{
							VariableIdentifier: variableIdentifier,
							TemplateFile:       template.Path,
							Line:               line,
						}
					}
				}
			}
		}

		line++
	}

	return nil
}

func (t *Templater) execute(template *TemplateFile, data any, targetPath string, allowNoValue bool) error {
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
		if err := t.checkUndefinedTemplateVariables(template, data); err != nil {
			return err
		}
	}

	err = WriteFile(t.outputFs, targetPath, out.Bytes(), template.Mode)
	if err != nil {
		return err
	}

	return nil
}

func (t *Templater) findTemplateVariableIdentifiers(templateData []byte) (variableIdentifiers []string) {
	// first, find all {{ .... }} strings in the line
	templateVariables := templateVariableRegex.FindAll(templateData, -1)
	for _, templateVariable := range templateVariables {

		// then, find all strings like ".foo.bar.baz" inside the {{ ... }}
		tplVariableIdentifier := templateVariableIdentifierRegex.FindAll(templateVariable, -1)
		for _, id := range tplVariableIdentifier {
			variableIdentifiers = append(variableIdentifiers, string(id))
		}
	}

	return variableIdentifiers
}

func (t *Templater) noValueDetection(template TemplateFile, renderedBytes string) error {
	scanner := bufio.NewScanner(strings.NewReader(renderedBytes))

	line := 1
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<no value>") {
			return fmt.Errorf("template '%s' uses variables with undefined value on line %d (line number is based on the rendered output and might not be accurate)", template.Path, line)
		}
		line++
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

func (t *Templater) getTemplateByPath(path string) *TemplateFile {
	for _, file := range t.Files {
		if file.Path == path {
			return file
		}
	}
	return nil
}
