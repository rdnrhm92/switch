package snowflake

import (
	"fmt"
	"sync"
)

// 全局业务管理器
var globalBusinessManager = NewBusinessManager()

// RegisterGlobalBusiness 注册全局业务
func RegisterGlobalBusiness(config BusinessConfig) error {
	return globalBusinessManager.RegisterBusiness(config)
}

// GenerateGlobalID 生成全局业务ID
func GenerateGlobalID(businessType BusinessType) (int64, error) {
	return globalBusinessManager.GenerateID(businessType)
}

// GenerateGlobalStringID 生成全局业务字符串ID
func GenerateGlobalStringID(businessType BusinessType) (string, error) {
	return globalBusinessManager.GeneratePrefixID(businessType)
}

// GenerateGlobalBatch 全局批量生成ID
func GenerateGlobalBatch(businessType BusinessType, count int) ([]int64, error) {
	return globalBusinessManager.GenerateBatch(businessType, count)
}

// GetGlobalBusinessTypes 获取全局已注册的业务类型
func GetGlobalBusinessTypes() []BusinessType {
	return globalBusinessManager.ListBusinessTypes()
}

// BusinessManager 业务管理器 用来管理多个业务 全局级别的
type BusinessManager struct {
	generators map[BusinessType]*BusinessGenerator
	mutex      sync.RWMutex
}

// NewBusinessManager 创建业务管理器
func NewBusinessManager() *BusinessManager {
	return &BusinessManager{
		generators: make(map[BusinessType]*BusinessGenerator),
	}
}

// RegisterBusiness 注册业务生成器
func (bm *BusinessManager) RegisterBusiness(config BusinessConfig) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if _, exists := bm.generators[config.BusinessType]; exists {
		return fmt.Errorf("business type '%s' already registered", config.BusinessType)
	}

	generator, err := NewBusinessGenerator(config)
	if err != nil {
		return fmt.Errorf("failed to create business generator: %w", err)
	}

	bm.generators[config.BusinessType] = generator
	return nil
}

// GenerateID 为指定业务生成ID
func (bm *BusinessManager) GenerateID(businessType BusinessType) (int64, error) {
	bm.mutex.RLock()
	generator, exists := bm.generators[businessType]
	bm.mutex.RUnlock()

	if !exists {
		return 0, fmt.Errorf("business type '%s' not registered", businessType)
	}

	return generator.GenerateID(), nil
}

// GeneratePrefixID 为指定业务生成前缀ID
func (bm *BusinessManager) GeneratePrefixID(businessType BusinessType) (string, error) {
	bm.mutex.RLock()
	generator, exists := bm.generators[businessType]
	bm.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("business type '%s' not registered", businessType)
	}

	return generator.GeneratePrefixID(), nil
}

// GenerateBatch 为指定业务批量生成ID
func (bm *BusinessManager) GenerateBatch(businessType BusinessType, count int) ([]int64, error) {
	bm.mutex.RLock()
	generator, exists := bm.generators[businessType]
	bm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("business type '%s' not registered", businessType)
	}

	return generator.GenerateBatch(count)
}

// GetGenerator 获取指定业务的生成器
func (bm *BusinessManager) GetGenerator(businessType BusinessType) (*BusinessGenerator, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	generator, exists := bm.generators[businessType]
	if !exists {
		return nil, fmt.Errorf("business type '%s' not registered", businessType)
	}

	return generator, nil
}

// ListBusinessTypes 列出所有已注册的业务类型
func (bm *BusinessManager) ListBusinessTypes() []BusinessType {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	types := make([]BusinessType, 0, len(bm.generators))
	for businessType := range bm.generators {
		types = append(types, businessType)
	}
	return types
}

// RemoveBusiness 移除业务生成器
func (bm *BusinessManager) RemoveBusiness(businessType BusinessType) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if _, exists := bm.generators[businessType]; !exists {
		return fmt.Errorf("business type '%s' not found", businessType)
	}

	delete(bm.generators, businessType)
	return nil
}
