package skipper_test

import (
	"bytes"
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	suite.Suite
	fileSystem afero.Fs
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}

func (suite *FileTestSuite) SetupTest() {
	suite.fileSystem = afero.NewMemMapFs()

	afero.WriteFile(suite.fileSystem, "/empty", []byte{}, 0644)
}

func (suite *FileTestSuite) TestLoadFile() {
	table := []struct {
		TestName      string
		ErrorExpected bool
		Path          string
	}{
		{
			TestName:      "EmptyFile",
			ErrorExpected: false,
			Path:          "/empty",
		},
		{
			TestName:      "NonExistingFile",
			ErrorExpected: true,
			Path:          "/unknownPath",
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {
			sut := skipper.File{
				Path: tt.Path,
			}
			err := sut.Load(suite.fileSystem)

			if tt.ErrorExpected {
				suite.Error(err, "Expected an error, but no error was returned")
			} else {
				suite.NoError(err, "Test should not return an error")
			}
		})
	}
}

type YamlFileTestSuite struct {
	suite.Suite
	fileSystem afero.Fs
}

func TestYamlFileTestSuite(t *testing.T) {
	suite.Run(t, new(YamlFileTestSuite))
}

func (suite *YamlFileTestSuite) SetupTest() {
	suite.fileSystem = afero.NewMemMapFs()

	afero.WriteFile(suite.fileSystem, "/empty", []byte{}, 0644)
}

func (suite *YamlFileTestSuite) TestNewYamlFile() {
	table := []struct {
		TestName      string
		ErrorExpected bool
		Path          string
	}{
		{
			TestName:      "EmptyPath",
			ErrorExpected: true,
			Path:          "",
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			file, err := skipper.NewYamlFile(tt.Path)

			if tt.ErrorExpected {
				suite.Assert().Error(err, "Expected an error, but no error was returned")
				suite.Assert().Nil(file)
			} else {
				suite.Assert().NoError(err, "Test should not return an error")
				suite.Assert().NotNil(file)
			}
		})
	}
}

func (suite *YamlFileTestSuite) TestLoad() {
	validYaml := `---
foo:
    bar: baz
    something: else`

	invalidYaml := `ThisIsReallyNotYaml`

	table := []struct {
		TestName      string
		ErrorExpected bool
		FileSystem    afero.Fs
		Path          string
		FileData      []byte
		ExpectedData  skipper.Data
	}{
		{
			TestName:      "NilFileSystem",
			ErrorExpected: true,
			FileSystem:    nil,
			Path:          "/empty",
			FileData:      nil,
			ExpectedData:  nil,
		},
		{
			TestName:      "EmptyFile",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/empty",
			FileData:      nil,
			ExpectedData:  nil,
		},
		{
			TestName:      "ValidYamlFile",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/validYaml",
			FileData:      []byte(validYaml),
			ExpectedData: skipper.Data{
				"foo": skipper.Data{
					"bar":       "baz",
					"something": "else",
				},
			},
		},
		{
			TestName:      "InvalidYamlFile",
			ErrorExpected: true,
			FileSystem:    suite.fileSystem,
			Path:          "/invalidYaml",
			FileData:      []byte(invalidYaml),
			ExpectedData:  nil,
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			// create proper YamlFile
			file, err := skipper.CreateNewYamlFile(suite.fileSystem, tt.Path, tt.FileData)
			suite.Assert().NotNil(file)
			suite.Assert().NoError(err)

			// test Load()
			err = file.Load(tt.FileSystem)

			if tt.ErrorExpected {
				suite.Assert().Error(err, "Expected an error, but no error was returned")
				suite.Assert().Nil(file.Data)
			} else {
				suite.Assert().NoError(err, "Test should not return an error")
				suite.Assert().Equal(tt.ExpectedData, file.Data)
			}
		})
	}
}

func (suite *YamlFileTestSuite) TestCreateFile() {
	table := []struct {
		TestName      string
		ErrorExpected bool
		FileSystem    afero.Fs
		Path          string
		Data          []byte
	}{
		{
			TestName:      "NilFileSystem",
			ErrorExpected: true,
			FileSystem:    nil,
			Path:          "",
			Data:          nil,
		},
		{
			TestName:      "EmptyPath",
			ErrorExpected: true,
			FileSystem:    suite.fileSystem,
			Path:          "",
			Data:          nil,
		},
		{
			TestName:      "CreateFile",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/newFile",
			Data:          []byte("hello world"),
		},
		{
			TestName:      "CreateFileAndPath",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/foo/bar/baz",
			Data:          []byte("hello world"),
		},
		// We're expecting to overwrite existing files, hence no error is expected.
		{
			TestName:      "CreateExistingFile",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/existing",
			Data:          []byte("hello world"),
		},
		{
			TestName:      "NilData",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/newFile",
			Data:          nil,
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			file, err := skipper.CreateNewYamlFile(tt.FileSystem, tt.Path, tt.Data)

			if tt.ErrorExpected {
				suite.Assert().Error(err, "Expected an error, but no error was returned")
				suite.Assert().Nil(file)
			} else {
				suite.Assert().NoError(err, "Test should not return an error")
				suite.Assert().NotNil(file)

				actualData, readErr := afero.ReadFile(suite.fileSystem, tt.Path)
				suite.Assert().NoError(readErr)

				// the 'NilData' test results in actualData being `[]byte{}`, and not `nil`, as expected
				// hence we assert that value, which is the same as returning nil.
				if tt.Data == nil {
					suite.Assert().Equal([]byte{}, actualData)
					return // test ends here for this case
				}

				suite.Assert().Equal(tt.Data, actualData)
			}
		})
	}
}

type TemplateFileTestSuite struct {
	suite.Suite
	fileSystem afero.Fs
}

func TestTemplateFileTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateFileTestSuite))
}

func (suite *TemplateFileTestSuite) SetupTest() {
	suite.fileSystem = afero.NewMemMapFs()
}

func (suite *TemplateFileTestSuite) TestNewTemplateFile() {
	table := []struct {
		TestName      string
		ErrorExpected bool
		Path          string
		Funcs         map[string]any
	}{
		{
			TestName:      "EmptyPath",
			ErrorExpected: true,
			Path:          "",
			Funcs:         nil,
		},
		{
			TestName:      "InvalidFuncs",
			ErrorExpected: true,
			Path:          "/filePath",
			Funcs: map[string]any{
				"foo": "bar",
			},
		},
		{
			TestName:      "ValidTemplateFile",
			ErrorExpected: false,
			Path:          "/filePath",
			Funcs: map[string]any{
				"foo": func() string { return "foo" },
			},
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			file, err := skipper.NewTemplateFile(tt.Path, tt.Funcs)

			if tt.ErrorExpected {
				suite.Assert().Error(err, "Expected an error, but no error was returned")
				suite.Assert().Nil(file)
			} else {
				suite.Assert().NoError(err, "Test should not return an error")
				suite.Assert().NotNil(file)
			}
		})
	}
}

func (suite *TemplateFileTestSuite) TestParse() {
	afero.WriteFile(suite.fileSystem, "/emptyFile", []byte(nil), 0644)

	table := []struct {
		TestName      string
		ErrorExpected bool
		FileSystem    afero.Fs
		Path          string
	}{
		{
			TestName:      "NonExistingFile",
			ErrorExpected: true,
			FileSystem:    suite.fileSystem,
			Path:          "/notExisting",
		},
		// Testing a 'valid template' is unnecessary because Parse() will only return errors if
		// an internal error in template.Parse occured (such as mutex errors).
		// An empty file is just as valid as any other file.
		{
			TestName:      "ParseEmptyFile",
			ErrorExpected: false,
			FileSystem:    suite.fileSystem,
			Path:          "/emptyFile",
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			file, err := skipper.NewTemplateFile(tt.Path, nil)
			suite.Assert().NotNil(file)
			suite.Assert().Nil(err)

			err = file.Parse(tt.FileSystem)

			if tt.ErrorExpected {
				suite.Assert().Error(err, "Expected an error, but no error was returned")
			} else {
				suite.Assert().NoError(err, "Test should not return an error")
			}
		})
	}
}

func (suite *TemplateFileTestSuite) TestExecute() {
	afero.WriteFile(suite.fileSystem, "/emptyFile", []byte(nil), 0644)
	afero.WriteFile(suite.fileSystem, "/templateVariable", []byte("{{ .Value }}"), 0644)

	table := []struct {
		TestName       string
		ErrorExpected  bool
		FileSystem     afero.Fs
		Path           string
		ContextData    any
		ExpectedOutput string
	}{
		{
			TestName:       "EmptyTemplate",
			ErrorExpected:  false,
			FileSystem:     suite.fileSystem,
			Path:           "/emptyFile",
			ContextData:    nil,
			ExpectedOutput: "",
		},
		{
			TestName:       "TemplateVariable",
			ErrorExpected:  false,
			FileSystem:     suite.fileSystem,
			Path:           "/templateVariable",
			ContextData:    map[string]interface{}{"Value": "hello"},
			ExpectedOutput: "hello",
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			file, err := skipper.NewTemplateFile(tt.Path, nil)
			suite.Assert().NotNil(file)
			suite.Assert().Nil(err)

			err = file.Parse(tt.FileSystem)
			suite.Assert().Nil(err)

			out := new(bytes.Buffer)
			err = file.Execute(out, tt.ContextData)

			if tt.ErrorExpected {
				suite.Assert().Error(err, "Expected an error, but no error was returned")
			} else {
				suite.Assert().NoError(err, "Test should not return an error")
				suite.Assert().Equal(tt.ExpectedOutput, out.String())
			}
		})
	}
}
