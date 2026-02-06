package repository

import (
	"context"
	"errors"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// ApprovalFormRepository 处理审批请求
type ApprovalFormRepository struct {
	BaseRepository
}

func NewApprovalFormRepository() *ApprovalFormRepository {
	return &ApprovalFormRepository{BaseRepository: BaseRepository{DB: GetDB()}}
}

func (r *ApprovalFormRepository) WithTx(tx *gorm.DB) *ApprovalFormRepository {
	return WithTx(r, tx)
}

// Create 创建一个新的审批请求
func (r *ApprovalFormRepository) Create(request *admin_model.Approval) error {
	return r.GetDB().Create(request).Error
}

// Update 更新审批请求
func (r *ApprovalFormRepository) Update(request *admin_model.Approval) error {
	return r.GetDB().Save(request).Error
}

// GetByID 根据ID获取审批请求
func (r *ApprovalFormRepository) GetByID(id uint) (*admin_model.Approval, error) {
	var request admin_model.Approval
	err := r.GetDB().First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetListByID 根据ID批量获取审批请求
func (r *ApprovalFormRepository) GetListByID(ids []uint, approvableType string) ([]*admin_model.Approval, error) {
	var requests []*admin_model.Approval
	err := r.GetDB().Where("id in ?", ids).Where("approvable_type = ?", approvableType).First(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// GetListByRequestUser 根据申请人批量获取审批请求
func (r *ApprovalFormRepository) GetListByRequestUser(id uint, approvableType string) ([]*admin_model.Approval, error) {
	var requests []*admin_model.Approval
	err := r.GetDB().Where("requester_user = ?", id).Where("approvable_type = ?", approvableType).Order("create_time desc").Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// applyCommonFilters 应用公共过滤条件
func (r *ApprovalFormRepository) applyCommonFilters(db *gorm.DB, req *dto.MyRequestReqBody) (*gorm.DB, error) {
	if req.Status != 0 {
		var status string
		switch req.Status {
		case 1:
			status = admin_model.ApprovalStatusPending
		case 2:
			status = admin_model.ApprovalStatusApproved
		case 3:
			status = admin_model.ApprovalStatusRejected
		default:
			return nil, errors.New("invalid status")
		}
		db = db.Where("status = ?", status)
	}

	if req.ApprovableType != 0 {
		var approvableType string
		switch req.ApprovableType {
		case 1:
			approvableType = admin_model.NamespaceType
		case 2:
			approvableType = admin_model.SwitchType
		default:
			return nil, errors.New("invalid approvableType")
		}
		db = db.Where("approvable_type = ?", approvableType)
	}

	if req.ApproverUser != 0 {
		db = db.Where("approver_user = ?", req.ApproverUser)
	}

	if req.RequesterUser != 0 {
		db = db.Where("requester_user = ?", req.RequesterUser)
	}

	return db, nil
}

// executeQuery 执行查询和计数
func (r *ApprovalFormRepository) executeQuery(db *gorm.DB) ([]*admin_model.Approval, int64, error) {
	var approvals []*admin_model.Approval
	var total int64

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return db.Session(&gorm.Session{}).Count(&total).Error
	})

	g.Go(func() error {
		return db.Session(&gorm.Session{}).Debug().Order("create_time desc").Find(&approvals).Error
	})

	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return approvals, total, nil
}

// Find 根据动态条件查询审批单
func (r *ApprovalFormRepository) Find(req *dto.MyRequestReqBody) ([]*admin_model.Approval, int64, error) {
	db := r.GetDB().Model(&admin_model.Approval{})

	// 应用用户身份过滤
	if req.Invited == 1 {
		db = db.Where("JSON_CONTAINS(approver_users, CAST(? AS JSON))", req.UId).Where("namespace_tag = ?", req.NamespaceTag)
	} else {
		db = db.Where("requester_user = ?", req.UId).Where("namespace_tag = ?", req.NamespaceTag)
	}

	// 应用公共过滤条件
	var err error
	db, err = r.applyCommonFilters(db, req)
	if err != nil {
		return nil, 0, err
	}

	return r.executeQuery(db)
}

// FindAll 查询所有审批单(不区分发起人跟受邀人)
func (r *ApprovalFormRepository) FindAll(req *dto.MyRequestReqBody) ([]*admin_model.Approval, int64, error) {
	db := r.GetDB().Model(&admin_model.Approval{})

	// 不区分发起人和受邀人，只按命名空间过滤
	if req.NamespaceTag != "" {
		db = db.Where("namespace_tag = ?", req.NamespaceTag)
	}

	// 应用公共过滤条件
	var err error
	db, err = r.applyCommonFilters(db, req)
	if err != nil {
		return nil, 0, err
	}

	return r.executeQuery(db)
}
