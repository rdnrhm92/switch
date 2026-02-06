package snowflake

import "fmt"

// BusinessGenerator 业务ID生成器
type BusinessGenerator struct {
	snowflake    *Generator
	businessType BusinessType
	prefix       string
}

// NewBusinessGenerator 创建业务ID生成器
func NewBusinessGenerator(config BusinessConfig) (*BusinessGenerator, error) {
	sf, err := NewGenerator(config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake: %w", err)
	}

	return &BusinessGenerator{
		snowflake:    sf,
		businessType: config.BusinessType,
		prefix:       config.Prefix,
	}, nil
}

// GenerateID 生成业务ID
func (bg *BusinessGenerator) GenerateID() int64 {
	return bg.snowflake.NextID()
}

// GeneratePrefixID 生成带前缀的字符串ID
func (bg *BusinessGenerator) GeneratePrefixID() string {
	id := bg.snowflake.NextID()

	if bg.prefix != "" {
		return fmt.Sprintf("%s%d", bg.prefix, id)
	}
	return fmt.Sprintf("%d", id)
}

// GenerateBatch 批量生成ID
func (bg *BusinessGenerator) GenerateBatch(count int) ([]int64, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		id := bg.snowflake.NextID()
		ids[i] = id
	}
	return ids, nil
}

// GenerateBatch 批量生成带前缀的ID
func (bg *BusinessGenerator) GeneratePrefixBatch(count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		id := bg.snowflake.NextID()
		ids[i] = fmt.Sprintf("%s%d", bg.prefix, id)
	}
	return ids, nil
}

// GetBusinessType 获取业务类型
func (bg *BusinessGenerator) GetBusinessType() BusinessType {
	return bg.businessType
}
