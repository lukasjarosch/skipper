package skipper_test

import (
	"fmt"
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewClass(t *testing.T) {
	tests := []struct {
		name            string
		namespace       skipper.Namespace
		dataSource      skipper.DataSource
		hasConfig       bool
		expectedConfig  *skipper.ClassConfig
		expectedRootKey string
		wantErr         bool
	}{
		{
			name:            "Valid input: Single level namespace; with valid config",
			namespace:       "test",
			dataSource:      new(mocks.DataSource),
			expectedRootKey: "test",
			hasConfig:       true,
			expectedConfig: &skipper.ClassConfig{
				Includes: []skipper.Namespace{
					skipper.Namespace("foo.bar"),
				},
			},
			wantErr: false,
		},
		{
			name:            "Valid input: Single level namespace; with invalid config",
			namespace:       "test",
			dataSource:      new(mocks.DataSource),
			expectedRootKey: "test",
			hasConfig:       true,
			expectedConfig:  nil,
			wantErr:         true,
		},
		{
			name:            "Valid input: Single level namespace; without config",
			namespace:       "test",
			dataSource:      new(mocks.DataSource),
			expectedRootKey: "test",
			hasConfig:       false,
			expectedConfig:  nil,
			wantErr:         false,
		},
		{
			name:            "Valid input: Single level namespace; without config",
			namespace:       "test",
			dataSource:      new(mocks.DataSource),
			expectedRootKey: "test",
			hasConfig:       false,
			expectedConfig:  nil,
			wantErr:         false,
		},
		{
			name:            "Valid input: Multi level namespace",
			namespace:       "test.namespace",
			dataSource:      new(mocks.DataSource),
			expectedRootKey: "namespace",
			hasConfig:       false,
			expectedConfig:  nil,
			wantErr:         false,
		},
		{
			name:           "Invalid input: empty namespace",
			namespace:      "",
			dataSource:     new(mocks.DataSource),
			hasConfig:      false,
			expectedConfig: nil,
			wantErr:        true,
		},
		{
			name:            "Valid input: Multi level namespace",
			namespace:       "test.namespace",
			dataSource:      new(mocks.DataSource),
			expectedRootKey: "namespace",
			hasConfig:       false,
			expectedConfig:  nil,
			wantErr:         false,
		},
		{
			name:           "Invalid input: empty namespace",
			namespace:      "",
			dataSource:     new(mocks.DataSource),
			hasConfig:      false,
			expectedConfig: nil,
			wantErr:        true,
		},
		{
			name:           "Invalid input: nil data source",
			namespace:      "test",
			dataSource:     nil,
			hasConfig:      false,
			expectedConfig: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// handle case where data is nil separately
			if tt.dataSource == nil && tt.wantErr {
				_, err := skipper.NewClass(tt.namespace, tt.dataSource)
				assert.Error(t, err)
				return
			}

			expectedClassRootKey := tt.namespace.Segments()[len(tt.namespace.Segments())-1]

			dataMock := (tt.dataSource).(*mocks.DataSource)
			dataMock.On("HasPath", skipper.DataPath{expectedClassRootKey, skipper.SkipperKey}).Return(tt.hasConfig)

			if tt.hasConfig {
				var ret error
				if tt.wantErr {
					ret = fmt.Errorf("unable to unmarshal path")
				} else {
					ret = nil
				}
				dataMock.On("UnmarshalPath", skipper.DataPath{expectedClassRootKey, skipper.SkipperKey}, &skipper.ClassConfig{}).Return(ret)
			}

			class, err := skipper.NewClass(tt.namespace, tt.dataSource)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			if tt.hasConfig {
				assert.NotNil(t, class.Configuration)
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.namespace, class.Namespace)
			assert.NotNil(t, class.Data)
			assert.Equal(t, tt.expectedRootKey, class.RootKey)
			dataMock.AssertExpectations(t)
		})
	}
}
