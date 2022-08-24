package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type ClassTestSuite struct {
	suite.Suite
	FileSystem      afero.Fs
	EmptyYamlFile   *skipper.YamlFile
	InvalidYamlFile *skipper.YamlFile
	ValidYamlFile   *skipper.YamlFile
}

func TestClassTestSuite(t *testing.T) {
	suite.Run(t, new(ClassTestSuite))
}

func (suite *ClassTestSuite) SetupTest() {
	suite.FileSystem = afero.NewMemMapFs()

	suite.EmptyYamlFile, _ = skipper.CreateNewYamlFile(suite.FileSystem, "/emptyYamlFile.yaml", []byte(nil))
}

func (suite *ClassTestSuite) TestNewClass() {
	noRootKey := suite.EmptyYamlFile
	noRootKey.Data = make(skipper.Data)

	multipleRootKeys := suite.EmptyYamlFile
	multipleRootKeys.Data = skipper.Data{
		"root1": skipper.Data{
			"foo": "bar",
		},
		"root2": skipper.Data{
			"bar": "baz",
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
			RelativeClassPath: "emptyYamlFile.yaml",
		},
		{
			TestName:          "DataNoRootKey",
			ErrorExpected:     true,
			YamlFile:          noRootKey,
			RelativeClassPath: "emptyYamlFile.yaml",
		},
		{
			TestName:          "MultipleRootKeys",
			ErrorExpected:     true,
			YamlFile:          multipleRootKeys,
			RelativeClassPath: "emptyYamlFile.yaml",
		},
		{
			TestName:          "ValidYamlFile",
			ErrorExpected:     true,
			YamlFile:          suite.ValidYamlFile,
			RelativeClassPath: "validYamlFile.yaml",
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
