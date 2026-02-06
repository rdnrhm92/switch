package repository

import (
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gorm.io/gorm"
)

// SwitchSnapshotConfigRepository 原子操作
type SwitchSnapshotConfigRepository struct {
	BaseRepository
}

func NewSwitchSnapshotConfigRepository() *SwitchSnapshotConfigRepository {
	return &SwitchSnapshotConfigRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (s *SwitchSnapshotConfigRepository) WithTx(tx *gorm.DB) *SwitchSnapshotConfigRepository {
	return WithTx(s, tx)
}

// Update 更新开关快照
func (s *SwitchSnapshotConfigRepository) Update(snapshot *admin_model.SwitchSnapshot) error {
	return s.GetDB().Save(snapshot).Error
}

// Create 创建开关快照
func (s *SwitchSnapshotConfigRepository) Create(snapshot *admin_model.SwitchSnapshot) error {
	return s.GetDB().Create(snapshot).Error
}

// FindByNamespaceAndEnvWithCursor 使用游标查询每个开关的最新版本快照
func (s *SwitchSnapshotConfigRepository) FindByNamespaceAndEnvWithCursor(namespaceTag, envTag string, lastSwitchID uint, pageSize int) ([]*admin_model.SwitchSnapshot, error) {
	var results []*admin_model.SwitchSnapshot

	// 关联子查询
	subQuery := s.GetDB().Table("switch_snapshots").
		Select("switch_id, MAX(version) as max_version").
		Where("namespace_tag = ? AND env_tag = ?", namespaceTag, envTag)

	// 子查询就使用游标提升效率
	if lastSwitchID > 0 {
		subQuery = subQuery.Where("switch_id > ?", lastSwitchID)
	}

	subQuery = subQuery.Group("switch_id").Limit(pageSize)

	db := s.GetDB().Table("switch_snapshots").
		Joins("INNER JOIN (?) as latest ON switch_snapshots.switch_id = latest.switch_id AND switch_snapshots.version = latest.max_version", subQuery).
		Where("switch_snapshots.namespace_tag = ? AND switch_snapshots.env_tag = ?", namespaceTag, envTag)

	// 按switch_id升序排序 保证游标分页的连续性
	err := db.Order("switch_snapshots.switch_id ASC").Scan(&results).Error
	return results, err
}
