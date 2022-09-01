package skipper_test

import (
	"math"
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/stretchr/testify/suite"
)

type DataTestSuite struct {
	suite.Suite
}

func TestDataTestSuite(t *testing.T) {
	suite.Run(t, new(DataTestSuite))
}

func (suite *DataTestSuite) TestNewData() {
	table := []struct {
		TestName      string
		ErrorExpected bool
		PanicExpected bool
		Input         interface{}
		ExpectedData  skipper.Data
	}{
		{
			TestName:      "DataFromSimpleStruct",
			ErrorExpected: false,
			Input: struct {
				Name string
			}{
				Name: "HelloWorld",
			},
			ExpectedData: skipper.Data{
				"name": "HelloWorld",
			},
		},
		{
			TestName:      "DataFromNestedStruct",
			ErrorExpected: false,
			Input: struct {
				Name       string
				Additional skipper.Data
			}{
				Name: "HelloWorld",
				Additional: skipper.Data{
					"Foo": "Baz",
				},
			},
			ExpectedData: skipper.Data{
				"name": "HelloWorld",
				"additional": skipper.Data{
					"Foo": "Baz",
				},
			},
		},
		{
			TestName:      "DataFromStructWithStructTags",
			ErrorExpected: false,
			Input: struct {
				Name       string       `yaml:"FancyName"`
				Additional skipper.Data `yaml:"EvenMoreData"`
			}{
				Name: "HelloWorld",
				Additional: skipper.Data{
					"Foo": "Baz",
				},
			},
			ExpectedData: skipper.Data{
				"FancyName": "HelloWorld",
				"EvenMoreData": skipper.Data{
					"Foo": "Baz",
				},
			},
		},
		{
			TestName:      "InvalidDataMarshalPanics",
			ErrorExpected: false,
			PanicExpected: true,
			Input:         func() {},
			ExpectedData:  nil,
		},
		{
			TestName:      "InvalidDataMarshalError",
			ErrorExpected: true,
			Input:         math.Inf(1),
			ExpectedData:  nil,
		},
		{
			TestName:      "InvalidDataUnmarshalError",
			ErrorExpected: true,
			Input:         []string{"foo", "bar"},
			ExpectedData:  nil,
		},
		{
			TestName:      "DataIsNil",
			ErrorExpected: false,
			Input:         nil,
			ExpectedData:  make(skipper.Data),
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {

			if tt.PanicExpected {
				suite.Panics(func() {
					_, _ = skipper.NewData(tt.Input)
				})
				return
			}

			data, err := skipper.NewData(tt.Input)

			if tt.ErrorExpected {
				suite.Error(err, "Expected an error, but no error was returned")
				suite.Nil(data)
			} else {
				suite.NoError(err, "Test should not return an error")
				suite.Equal(tt.ExpectedData, data)
			}
		})
	}
}

func (suite *DataTestSuite) TestHasKey() {
	table := []struct {
		TestName       string
		Data           skipper.Data
		Key            string
		ExpectedResult bool
	}{
		{
			TestName:       "KeyDoesNotExist",
			Data:           skipper.Data{},
			Key:            "nothing",
			ExpectedResult: false,
		},
		{
			TestName: "EmptyKey",
			Data: skipper.Data{
				"something": "bar",
			},
			Key:            "",
			ExpectedResult: false,
		},
		{
			TestName: "KeyExists",
			Data: skipper.Data{
				"something": "bar",
			},
			Key:            "something",
			ExpectedResult: true,
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {
			res := tt.Data.HasKey(tt.Key)
			suite.Equal(tt.ExpectedResult, res)
		})
	}
}

func (suite *DataTestSuite) TestGet() {
	table := []struct {
		TestName       string
		Data           skipper.Data
		Key            string
		ExpectedResult skipper.Data
	}{
		{
			TestName:       "KeyDoesNotExist",
			Data:           skipper.Data{},
			Key:            "nothing",
			ExpectedResult: skipper.Data(nil),
		},
		{
			TestName: "KeyExists",
			Data: skipper.Data{
				"something": "bar",
			},
			Key:            "something",
			ExpectedResult: nil,
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {
			res := tt.Data.Get(tt.Key)
			suite.Equal(tt.ExpectedResult, res)
		})
	}
}

func (suite *DataTestSuite) TestGetPath() {
	validData := skipper.Data{
		"RootKey": skipper.Data{
			"Name": "Peter",
			"Additional": skipper.Data{
				"Foo": "Bar",
				"Baz": []interface{}{"one", "two", "three"},
			},
		},
	}
	invalidData := skipper.Data{
		"RootKey": skipper.Data{
			"Name": []string{"Peter", "Max"},
		},
	}

	table := []struct {
		TestName       string
		ErrorExpected  bool
		ErrorSubstring string
		Data           skipper.Data
		Path           []interface{}
		ExpectedResult interface{}
	}{
		{
			TestName:       "KeyNotFoundInData",
			ErrorExpected:  true,
			ErrorSubstring: "key not found",
			Data:           validData,
			Path:           []interface{}{"foo", "bar"},
			ExpectedResult: nil,
		},
		{
			TestName:       "InvalidPathForDataMap",
			ErrorExpected:  true,
			ErrorSubstring: "unexpected key type (int) for Data",
			Data:           validData,
			Path:           []interface{}{"RootKey", 42},
			ExpectedResult: nil,
		},
		{
			TestName:       "ValidPathForDataMap",
			ErrorExpected:  false,
			ErrorSubstring: "",
			Data:           validData,
			Path:           []interface{}{"RootKey", "Name"},
			ExpectedResult: "Peter",
		},
		{
			TestName:       "InvalidPathForSlice",
			ErrorExpected:  true,
			ErrorSubstring: "unexpected key type (string) for []interface{}",
			Data:           validData,
			Path:           []interface{}{"RootKey", "Additional", "Baz", "INVALID"},
			ExpectedResult: nil,
		},
		{
			TestName:       "ValidPathForSlice",
			ErrorExpected:  false,
			ErrorSubstring: "",
			Data:           validData,
			Path:           []interface{}{"RootKey", "Additional", "Baz", 2},
			ExpectedResult: "three",
		},
		{
			TestName:       "IndexOutOfRangeForSlice",
			ErrorExpected:  true,
			ErrorSubstring: "index out of range: 42",
			Data:           validData,
			Path:           []interface{}{"RootKey", "Additional", "Baz", 42},
			ExpectedResult: nil,
		},
		{
			TestName:       "InvalidDataAccess",
			ErrorExpected:  true,
			ErrorSubstring: "unexpected node type",
			Data:           invalidData,
			Path:           []interface{}{"RootKey", "Name", 2},
			ExpectedResult: nil,
		},
	}

	for _, tt := range table {
		suite.Run(tt.TestName, func() {
			res, err := tt.Data.GetPath(tt.Path...)

			if tt.ErrorExpected {
				suite.Error(err, "Expected an error, but no error was returned")
				suite.ErrorContains(err, tt.ErrorSubstring, "Unexpected error string returned")
				suite.Nil(res)
			} else {
				suite.NoError(err, "Test should not return an error")
				suite.Equal(tt.ExpectedResult, res)
			}
		})
	}
}
