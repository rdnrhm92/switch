package snowflake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  int64
		wantErr bool
	}{
		{
			name:    "valid node ID",
			nodeID:  1,
			wantErr: false,
		},
		{
			name:    "invalid node ID (too large)",
			nodeID:  1024,
			wantErr: true,
		},
		{
			name:    "invalid node ID (negative)",
			nodeID:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewGenerator(tt.nodeID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)
			}
		})
	}
}

func TestGenerator_NextID(t *testing.T) {
	generator, err := NewGenerator(1)
	assert.NoError(t, err)

	// Generate multiple IDs and verify they are unique
	idMap := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id := generator.NextID()
		assert.False(t, idMap[id], "Generated ID should be unique")
		idMap[id] = true
	}
}

func TestGenerator_NextIDString(t *testing.T) {
	generator, err := NewGenerator(1)
	assert.NoError(t, err)

	// Generate multiple string IDs and verify they are unique
	idMap := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generator.NextIDString()
		assert.False(t, idMap[id], "Generated string ID should be unique")
		idMap[id] = true
	}
}

func TestGetDefault(t *testing.T) {
	// Get default generator multiple times
	gen1 := GetDefault()
	gen2 := GetDefault()

	// Verify it's the same instance
	assert.Same(t, gen1, gen2, "Default generator should be singleton")

	// Verify it works
	id := gen1.NextID()
	assert.Greater(t, id, int64(0), "Generated ID should be positive")
}
