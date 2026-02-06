package snowflake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBusinessManager_RegisterBusiness(t *testing.T) {
	manager := NewBusinessManager()

	tests := []struct {
		name    string
		config  BusinessConfig
		wantErr bool
	}{
		{
			name: "register new business",
			config: BusinessConfig{
				BusinessType: "order",
				MachineID:    1,
				Prefix:       "ORD-",
			},
			wantErr: false,
		},
		{
			name: "register duplicate business",
			config: BusinessConfig{
				BusinessType: "order",
				MachineID:    2,
				Prefix:       "ORD-",
			},
			wantErr: true,
		},
		{
			name: "register with invalid machine ID",
			config: BusinessConfig{
				BusinessType: "payment",
				MachineID:    1024,
				Prefix:       "PAY-",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RegisterBusiness(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBusinessManager_GenerateID(t *testing.T) {
	manager := NewBusinessManager()
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	// Register business
	err := manager.RegisterBusiness(config)
	assert.NoError(t, err)

	// Test generating ID for registered business
	id, err := manager.GenerateID("order")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Test generating ID for unregistered business
	_, err = manager.GenerateID("unknown")
	assert.Error(t, err)
}

func TestBusinessManager_GeneratePrefixID(t *testing.T) {
	manager := NewBusinessManager()
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	// Register business
	err := manager.RegisterBusiness(config)
	assert.NoError(t, err)

	// Test generating prefix ID for registered business
	id, err := manager.GeneratePrefixID("order")
	assert.NoError(t, err)
	assert.Regexp(t, "^ORD-\\d+$", id)

	// Test generating prefix ID for unregistered business
	_, err = manager.GeneratePrefixID("unknown")
	assert.Error(t, err)
}

func TestBusinessManager_GenerateBatch(t *testing.T) {
	manager := NewBusinessManager()
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	// Register business
	err := manager.RegisterBusiness(config)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		businessType BusinessType
		count        int
		wantErr      bool
	}{
		{
			name:         "valid batch for registered business",
			businessType: "order",
			count:        100,
			wantErr:      false,
		},
		{
			name:         "invalid count",
			businessType: "order",
			count:        0,
			wantErr:      true,
		},
		{
			name:         "unregistered business",
			businessType: "unknown",
			count:        100,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := manager.GenerateBatch(tt.businessType, tt.count)
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

func TestBusinessManager_GetGenerator(t *testing.T) {
	manager := NewBusinessManager()
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	// Register business
	err := manager.RegisterBusiness(config)
	assert.NoError(t, err)

	// Test getting registered generator
	generator, err := manager.GetGenerator("order")
	assert.NoError(t, err)
	assert.NotNil(t, generator)
	assert.Equal(t, BusinessType("order"), generator.GetBusinessType())

	// Test getting unregistered generator
	generator, err = manager.GetGenerator("unknown")
	assert.Error(t, err)
	assert.Nil(t, generator)
}

func TestBusinessManager_ListBusinessTypes(t *testing.T) {
	manager := NewBusinessManager()
	configs := []BusinessConfig{
		{
			BusinessType: "order",
			MachineID:    1,
			Prefix:       "ORD-",
		},
		{
			BusinessType: "payment",
			MachineID:    2,
			Prefix:       "PAY-",
		},
	}

	// Register businesses
	for _, config := range configs {
		err := manager.RegisterBusiness(config)
		assert.NoError(t, err)
	}

	// Test listing business types
	types := manager.ListBusinessTypes()
	assert.Len(t, types, len(configs))
	assert.Contains(t, types, BusinessType("order"))
	assert.Contains(t, types, BusinessType("payment"))
}

func TestBusinessManager_RemoveBusiness(t *testing.T) {
	manager := NewBusinessManager()
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}

	// Register business
	err := manager.RegisterBusiness(config)
	assert.NoError(t, err)

	// Test removing registered business
	err = manager.RemoveBusiness("order")
	assert.NoError(t, err)

	// Verify business was removed
	_, err = manager.GetGenerator("order")
	assert.Error(t, err)

	// Test removing unregistered business
	err = manager.RemoveBusiness("unknown")
	assert.Error(t, err)
}

func TestGlobalBusinessManager(t *testing.T) {
	// Test registering global business
	config := BusinessConfig{
		BusinessType: "order",
		MachineID:    1,
		Prefix:       "ORD-",
	}
	err := RegisterGlobalBusiness(config)
	assert.NoError(t, err)

	// Test generating global ID
	id, err := GenerateGlobalID("order")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Test generating global string ID
	strID, err := GenerateGlobalStringID("order")
	assert.NoError(t, err)
	assert.Regexp(t, "^ORD-\\d+$", strID)

	// Test generating global batch
	ids, err := GenerateGlobalBatch("order", 100)
	assert.NoError(t, err)
	assert.Len(t, ids, 100)

	// Test getting global business types
	types := GetGlobalBusinessTypes()
	assert.Contains(t, types, BusinessType("order"))
}
