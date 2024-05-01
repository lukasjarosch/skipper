package data_test

import (
	"testing"
	"time"

	. "github.com/lukasjarosch/skipper/v1/data"
	"github.com/stretchr/testify/assert"
)

func TestValueString(t *testing.T) {
	testCases := []struct {
		name     string
		value    Value
		expected string
	}{
		{"IntegerValueToString", NewValue(42), "42"},
		{"StringValueToString", NewValue("hello"), "hello"},
		{"NilValueToString", NewValue(nil), "<nil>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValueMap(t *testing.T) {
	testCases := []struct {
		name     string
		value    Value
		expected map[string]interface{}
		err      bool
	}{
		{"ValidMapValue", NewValue(map[string]interface{}{"key": "value"}), map[string]interface{}{"key": "value"}, false},
		{"InvalidMapValue", NewValue(42), nil, true},
		{"NilValue", NewValue(nil), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.value.Map()
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestValueDuration(t *testing.T) {
	testCases := []struct {
		name     string
		value    Value
		expected time.Duration
		err      bool
	}{
		{"ValidDurationString", NewValue("1s"), time.Second, false},
		{"InvalidDurationString", NewValue("invalid_duration"), 0, true},
		{"NilValue", NewValue(nil), 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.value.Duration()
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestValueInt(t *testing.T) {
	testCases := []struct {
		name     string
		value    Value
		expected int
		err      bool
	}{
		{"ValidIntValue", NewValue(42), 42, false},
		{"InvalidIntValue", NewValue("invalid_int"), 0, true},
		{"NilValue", NewValue(nil), 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.value.Int()
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
