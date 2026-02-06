package snowflake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBusinessGenerator(t *testing.T) {
	tests := []struct {
		name    string
		config  BusinessConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: BusinessConfig{
				BusinessType: "order",
				MachineID:    1,
				Prefix:       "ORD-",
			},
			wantErr: false,
		},
		{
			name: "invalid machine ID",
			config: BusinessConfig{
				BusinessType: "order",
				MachineID:    1024,
				Prefix:       "ORD-",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewBusinessGenerator(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)
				assert.Equal(t, tt.config.BusinessType, generator.GetBusinessType())
			}
		})
	}
}

func TestBusinessGenerator_GenerateID(t *testing.T) {
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	generator, err := NewBusinessGenerator(config)
	assert.NoError(t, err)

	// Generate multiple IDs and verify they are unique
	idMap := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id := generator.GenerateID()
		assert.False(t, idMap[id], "Generated ID should be unique")
		idMap[id] = true
	}
}

func TestBusinessGenerator_GeneratePrefixID(t *testing.T) {
	tests := []struct {
		name      string
		config    BusinessConfig
		wantRegex string
	}{
		{
			name: "with prefix",
			config: BusinessConfig{
				BusinessType: "order",
				MachineID:    1,
				Prefix:       "ORD-",
			},
			wantRegex: "^ORD-\\d+$",
		},
		{
			name: "without prefix",
			config: BusinessConfig{
				BusinessType: "order",
				MachineID:    1,
				Prefix:       "",
			},
			wantRegex: "^\\d+$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewBusinessGenerator(tt.config)
			assert.NoError(t, err)

			// Generate multiple IDs and verify they are unique and match the pattern
			idMap := make(map[string]bool)
			for i := 0; i < 1000; i++ {
				id := generator.GeneratePrefixID()
				assert.False(t, idMap[id], "Generated ID should be unique")
				assert.Regexp(t, tt.wantRegex, id)
				idMap[id] = true
			}
		})
	}
}

func TestBusinessGenerator_GenerateBatch(t *testing.T) {
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	generator, err := NewBusinessGenerator(config)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{
			name:    "valid count",
			count:   100,
			wantErr: false,
		},
		{
			name:    "zero count",
			count:   0,
			wantErr: true,
		},
		{
			name:    "negative count",
			count:   -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := generator.GenerateBatch(tt.count)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ids)
			} else {
				assert.NoError(t, err)
				assert.Len(t, ids, tt.count)

				// Verify uniqueness
				idMap := make(map[int64]bool)
				for _, id := range ids {
					assert.False(t, idMap[id], "Generated ID should be unique")
					idMap[id] = true
				}
			}
		})
	}
}

func TestBusinessGenerator_GeneratePrefixBatch(t *testing.T) {
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	generator, err := NewBusinessGenerator(config)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		count     int
		wantErr   bool
		wantRegex string
	}{
		{
			name:      "valid count",
			count:     100,
			wantErr:   false,
			wantRegex: "^ORD-\\d+$",
		},
		{
			name:    "zero count",
			count:   0,
			wantErr: true,
		},
		{
			name:    "negative count",
			count:   -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := generator.GeneratePrefixBatch(tt.count)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ids)
			} else {
				assert.NoError(t, err)
				assert.Len(t, ids, tt.count)

				// Verify uniqueness and pattern
				idMap := make(map[string]bool)
				for _, id := range ids {
					assert.False(t, idMap[id], "Generated ID should be unique")
					assert.Regexp(t, tt.wantRegex, id)
					idMap[id] = true
				}
			}
		})
	}
}
