package repository

import (
	"context"
	"log"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_driver"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDB 从驱动管理器获取数据库连接
func GetDB() *gorm.DB {
	drv, err := admin_driver.GetMysql("default")
	if err != nil {
		log.Fatalf("failed to get mysql admin_driver: %v", err)
		return nil
	}
	return drv.Db()
}

func FindWithSuperAdmin(ctx context.Context, db *gorm.DB) *gorm.DB {
	return findWith(ctx, db, func(db *gorm.DB, userInfo *dto.UserInfo) *gorm.DB {
		return db.Where("namespace.tag = ? ", userInfo.SelectNamespace)
	})
}

func FindWithSuperAdminV2(ctx context.Context, db *gorm.DB) *gorm.DB {
	return findWith(ctx, db, func(db *gorm.DB, userInfo *dto.UserInfo) *gorm.DB {
		return db.Joins("JOIN namespace_members ON namespace_members.user_id = users.id").
			Where("namespace_members.namespace_tag = ?", userInfo.SelectNamespace).
			Distinct()
	})
}

func FindSwitchFactorWithSuperAdmin(ctx context.Context, db *gorm.DB) *gorm.DB {
	return findWith(ctx, db, func(db *gorm.DB, userInfo *dto.UserInfo) *gorm.DB {
		return db.Where("switch_factor.namespace_tag in (?) ", []string{userInfo.SelectNamespace, ""})
	})
}

func FindSwitchWithSuperAdmin(ctx context.Context, db *gorm.DB) *gorm.DB {
	return findWith(ctx, db, func(db *gorm.DB, userInfo *dto.UserInfo) *gorm.DB {
		return db.Where("switches.namespace_tag = ? ", userInfo.SelectNamespace)
	})
}

func findWith(ctx context.Context, db *gorm.DB, selCondition func(*gorm.DB, *dto.UserInfo) *gorm.DB) *gorm.DB {
	userInfo := info.GetUserInfo(ctx)
	if userInfo == nil {
		return selCondition(db, userInfo)
	}
	if userInfo.SelectNamespace == "" {
		return db.Where("1 = 0")
	}
	return selCondition(db, userInfo)
}

func FindWithSuperAdminWithJoin(ctx *gin.Context, db *gorm.DB) *gorm.DB {
	return findWith(ctx, db, func(db *gorm.DB, userInfo *dto.UserInfo) *gorm.DB {
		return db.Where("environments.namespace_tag = ?", userInfo.SelectNamespace)
	})
}
