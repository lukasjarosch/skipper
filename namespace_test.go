package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/stretchr/testify/assert"
)

func TestNewNamespace(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		expectedNS  skipper.Namespace
		expectedErr error
	}{
		{
			name:        "Valid input",
			namespace:   "test.namespace",
			expectedNS:  skipper.Namespace("test.namespace"),
			expectedErr: nil,
		},
		{
			name:        "Empty namespace",
			namespace:   "",
			expectedNS:  skipper.Namespace(""),
			expectedErr: skipper.ErrEmptyNamespace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, err := skipper.NewNamespace(tt.namespace)
			assert.Equal(t, tt.expectedNS, ns)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestNamespace_Segments(t *testing.T) {
	ns, _ := skipper.NewNamespace("test.namespace")
	expected := []string{"test", "namespace"}
	assert.Equal(t, expected, ns.Segments())
}

func TestNamespace_String(t *testing.T) {
	ns, _ := skipper.NewNamespace("test.namespace")
	expected := "test.namespace"
	assert.Equal(t, expected, ns.String())
}
