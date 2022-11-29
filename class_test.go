package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type ClassTestSuite struct {
	suite.Suite
	FileSystem          afero.Fs
	EmptyYamlFile       *skipper.YamlFile
	InvalidYamlFile     *skipper.YamlFile
	TooManyKeysYamlFile *skipper.YamlFile
	ValidYamlFile       *skipper.YamlFile
}

func TestClassTestSuite(t *testing.T) {
	suite.Run(t, new(ClassTestSuite))
}

func (suite *ClassTestSuite) SetupTest() {
	suite.FileSystem = afero.NewMemMapFs()

	suite.EmptyYamlFile, _ = skipper.CreateNewYamlFile(suite.FileSystem, "/emptyYaml.yaml", []byte(nil))
	suite.ValidYamlFile, _ = skipper.CreateNewYamlFile(suite.FileSystem, "/validYaml.yaml", []byte(nil))
	suite.InvalidYamlFile, _ = skipper.CreateNewYamlFile(suite.FileSystem, "/invalidYaml.yaml", []byte(nil))
	suite.TooManyKeysYamlFile, _ = skipper.CreateNewYamlFile(suite.FileSystem, "/tooManyKeys.yaml", []byte(nil))
}

func (suite *ClassTestSuite) TestNewClass() {
	noRootKey := suite.EmptyYamlFile
	noRootKey.Data = make(skipper.Data)

	rootKeyFalse := suite.InvalidYamlFile
	rootKeyFalse.Data = skipper.Data{
		"asdasd": "test",
	}

	multipleRootKeys := suite.TooManyKeysYamlFile
	multipleRootKeys.Data = skipper.Data{
		"root1": skipper.Data{
			"foo": "bar",
		},
		"root2": skipper.Data{
			"bar": "baz",
		},
	}

	validYaml := suite.ValidYamlFile
	validYaml.Data = skipper.Data{
		"validYaml": skipper.Data{
			"foo": "bar",
		},
	}

	table := []struct {
		TestName          string
		ErrorExpected     bool
		YamlFile          *skipper.YamlFile
		RelativeClassPath string
	}{
		{
			TestName:          "YamlFileNil",
			ErrorExpected:     true,
			YamlFile:          nil,
			RelativeClassPath: "",
		},
		{
			TestName:          "EmptyRelativeClassPath",
			ErrorExpected:     true,
			YamlFile:          suite.EmptyYamlFile,
			RelativeClassPath: "",
		},
		{
			TestName:          "EmptyYamlFile",
			ErrorExpected:     true,
			YamlFile:          suite.EmptyYamlFile,
			RelativeClassPath: "emptyYaml.yaml",
		},
		{
			TestName:          "DataNoRootKey",
			ErrorExpected:     true,
			YamlFile:          noRootKey,
			RelativeClassPath: "emptyYaml.yaml",
		},
		{
			TestName:          "MultipleRootKeys",
			ErrorExpected:     true,
			YamlFile:          suite.TooManyKeysYamlFile,
			RelativeClassPath: "tooManyKeys.yaml",
		},
		{
			TestName:          "RootKeyDoesNotMatchYamlFileName",
			ErrorExpected:     true,
			YamlFile:          rootKeyFalse,
			RelativeClassPath: "invalidYaml.yaml",
		},
		{
			TestName:          "ValidYamlFile",
			ErrorExpected:     false,
			YamlFile:          validYaml,
			RelativeClassPath: "validYaml.yaml",
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {
			class, err := skipper.NewClass(tt.YamlFile, tt.RelativeClassPath)

			if tt.ErrorExpected {
				suite.Error(err, "Expected an error, but no error was returned")
				suite.Nil(class)
			} else {
				suite.NoError(err, "Test should not return an error")
				suite.NotNil(class)
			}
		})
	}
}

func (suite *ClassTestSuite) TestRootKey() {
	table := []struct {
		TestName        string
		Data            skipper.Data
		ExpectedRootKey string
	}{
		{
			TestName:        "EmptyData",
			Data:            make(skipper.Data),
			ExpectedRootKey: "",
		},
		{
			TestName: "SingleRootKey",
			Data: skipper.Data{
				"foo": skipper.Data{
					"bar": "baz",
				},
			},
			ExpectedRootKey: "foo",
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			sut := &skipper.Class{
				File: &skipper.YamlFile{
					File: skipper.File{
						Path:  "unused",
						Bytes: []byte{},
					},
					Data: tt.Data,
				},
				Name: "unused",
			}

			rootKey := sut.RootKey()
			suite.Equal(tt.ExpectedRootKey, rootKey)
		})
	}
}

func (suite *ClassTestSuite) TestNameAsIdentifier() {
	table := []struct {
		TestName           string
		ClassName          string
		ExpectedIdentifier []interface{}
	}{
		{
			TestName:           "EmptyName",
			ClassName:          "",
			ExpectedIdentifier: []interface{}{""},
		},
		{
			TestName:           "SingleSeparatedName",
			ClassName:          "Foo.Bar",
			ExpectedIdentifier: []interface{}{"Foo", "Bar"},
		},
		{
			TestName:           "MultipleSeparatedName",
			ClassName:          "Foo.Bar.Baz.Hello",
			ExpectedIdentifier: []interface{}{"Foo", "Bar", "Baz", "Hello"},
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			sut := &skipper.Class{
				Name: tt.ClassName,
			}

			actual := sut.NameAsIdentifier()
			suite.Equal(tt.ExpectedIdentifier, actual)
		})
	}
}
