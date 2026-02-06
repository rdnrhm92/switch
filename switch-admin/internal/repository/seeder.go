package repository

import (
	"fmt"

	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-sdk-core/factor"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const createBy = "系统"

func SeedData(db *gorm.DB) {
	permissions := []admin_model.Permission{
		// 命名空间权限
		{Name: "namespaces:edit", Description: "编辑命名空间", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 用户权限
		{Name: "users:watch", Description: "查看用户列表", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "users:list-for-switch", Description: "查看用户列表(选择审批人)", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "users:assign-roles", Description: "分配用户角色", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 角色权限
		{Name: "roles:watch", Description: "查看角色列表", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "roles:create", Description: "创建角色", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "roles:edit", Description: "编辑角色", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 权限管理
		{Name: "permissions:watch", Description: "查看权限列表", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "permissions:create", Description: "创建权限", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "permissions:edit", Description: "编辑权限", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 环境权限
		{Name: "namespace-envs:watch", Description: "查看命名空间环境(选择环境)", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "envs:watch", Description: "查看环境列表", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "envs:create", Description: "创建环境", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "envs:edit", Description: "编辑环境", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "envs:publish", Description: "发布环境", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 开关权限
		{Name: "switches:watch", Description: "查看开关列表", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "switches:create", Description: "创建开关", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "switches:edit", Description: "编辑开关", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "switches:push", Description: "推送开关", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 因子信息权限
		{Name: "switch-factors:watch", Description: "查看因子信息", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "switch-factors:create", Description: "创建因子信息", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "switch-factors:edit", Description: "编辑因子信息", CommonModel: model.CommonModel{CreatedBy: createBy}},

		// 审批权限
		{Name: "approvals:watch", Description: "查看审批单", CommonModel: model.CommonModel{CreatedBy: createBy}},
		{Name: "approvals:approve", Description: "审批", CommonModel: model.CommonModel{CreatedBy: createBy}},
	}

	for _, p := range permissions {
		db.Attrs(admin_model.Permission{Description: p.Description, CommonModel: model.CommonModel{CreatedBy: createBy}}).FirstOrCreate(&p, admin_model.Permission{Name: p.Name})
	}
	fmt.Println("Permissions seeded.")

	AdminRole.Name = "Admin"
	db.Attrs(admin_model.Role{Description: "管理员，拥有命名空间下的所有权限", CommonModel: model.CommonModel{CreatedBy: createBy}, NamespaceTag: ""}).FirstOrCreate(AdminRole, admin_model.Role{Name: "Admin"})

	DeveloperRole.Name = "Developer"
	db.Attrs(admin_model.Role{Description: "开发者，可以管理开关、审批，以及查看相关资源，但无对空间环境用户等操作权限", CommonModel: model.CommonModel{CreatedBy: createBy}, NamespaceTag: ""}).FirstOrCreate(DeveloperRole, admin_model.Role{Name: "Developer"})

	ObserverRole.Name = "Observer"
	db.Attrs(admin_model.Role{Description: "观察者，拥有所有资源的查看权限，但无操作权限", CommonModel: model.CommonModel{CreatedBy: createBy}, NamespaceTag: ""}).FirstOrCreate(ObserverRole, admin_model.Role{Name: "Observer"})

	OrdinaryRole.Name = "Ordinary"
	db.Attrs(admin_model.Role{Description: "普通用户，拥有基础的查看权限", CommonModel: model.CommonModel{CreatedBy: createBy}, NamespaceTag: ""}).FirstOrCreate(OrdinaryRole, admin_model.Role{Name: "Ordinary"})

	fmt.Println("Roles seeded.")

	assignAllPermissionsToAdmin(db, AdminRole)
	assignDeveloperPermissions(db, DeveloperRole)
	assignObserverPermissions(db, ObserverRole)
	assignOrdinaryPermissions(db, OrdinaryRole)
	fmt.Println("Role-Permission assignments seeded.")

	// 创建一个默认的super_admin用户
	createDefaultAdminUser(db)
	// 创建因子信息
	addSwitchFactor(db, "")
}

func createDefaultAdminUser(db *gorm.DB) {
	if err := db.Where("username = ?", "admin").First(SuperAdmin).Error; err == nil {
		logger.Logger.Info("Admin user already exists.")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("factor"), bcrypt.DefaultCost)
	if err != nil {
		logger.Logger.Errorf("Failed to hash admin password: %v", err)
		return
	}
	SuperAdmin.Username = "admin"
	SuperAdmin.Password = string(hashedPassword)
	SuperAdmin.CommonModel = model.CommonModel{CreatedBy: createBy}
	//超级管理员拥有所有租户的所有权限
	isTrue := true
	SuperAdmin.IsSuperAdmin = &isTrue

	userRepo := NewUserRepository()
	//超级管理员不走权限角色那套
	if err = userRepo.Create(SuperAdmin); err != nil {
		logger.Logger.Infof("Admin user created has error [%s]", err)
	}
}

func assignAllPermissionsToAdmin(db *gorm.DB, adminRole *admin_model.Role) {
	// 所有的权限
	var allPermissions []admin_model.Permission
	db.Where("permissions.namespace_tag = ''").Find(&allPermissions)
	for _, p := range allPermissions {
		db.Attrs(admin_model.RolePermission{CommonModel: model.CommonModel{CreatedBy: createBy}}).FirstOrCreate(&admin_model.RolePermission{}, &admin_model.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID})
	}
}

func assignDeveloperPermissions(db *gorm.DB, devRole *admin_model.Role) {
	permissionsToAssign := []string{
		"switches:watch",
		"switches:create",
		"switches:edit",
		"switches:push",
		"users:list-for-switch", // 联动权限
		"namespace-envs:watch",  // 联动权限
		"switch-factors:watch",  // 联动权限
		"approvals:watch",
		"approvals:approve",
		"users:watch",
		"roles:watch",
		"permissions:watch",
		"envs:watch",
	}
	assignPermissions(db, devRole, permissionsToAssign)
}

func assignObserverPermissions(db *gorm.DB, observerRole *admin_model.Role) {
	permissionsToAssign := []string{
		"users:watch",
		"roles:watch",
		"permissions:watch",
		"envs:watch",
		"namespace-envs:watch",
		"switches:watch",
		"switch-factors:watch",
		"approvals:watch",
	}
	assignPermissions(db, observerRole, permissionsToAssign)
}

func assignOrdinaryPermissions(db *gorm.DB, ordinaryRole *admin_model.Role) {
	permissionsToAssign := []string{
		"switches:watch",
		"namespace-envs:watch", // 联动权限
		"switch-factors:watch",
		"approvals:watch",
	}
	assignPermissions(db, ordinaryRole, permissionsToAssign)
}

func assignPermissions(db *gorm.DB, role *admin_model.Role, permissionNames []string) {
	for _, pName := range permissionNames {
		var p admin_model.Permission
		if db.Where("name = ?", pName).First(&p).Error == nil {
			db.Attrs(admin_model.RolePermission{CommonModel: model.CommonModel{CreatedBy: createBy}}).FirstOrCreate(&admin_model.RolePermission{}, &admin_model.RolePermission{RoleID: role.ID, PermissionID: p.ID})
		}
	}
}

func addSwitchFactor(tx *gorm.DB, nsTag string) {
	metas := []*admin_model.SwitchFactor{
		{
			Factor:       factor.Custom_Whole.Factor,
			Name:         "非开放自定义开关",
			Description:  factor.Custom_Whole.Description,
			JsonSchema:   factor.Custom_Whole.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.Custom_Arbitrarily.Factor,
			Name:         "开放自定义开关",
			Description:  factor.Custom_Arbitrarily.Description,
			JsonSchema:   factor.Custom_Arbitrarily.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.UserNick.Factor,
			Name:         "昵称开关",
			Description:  factor.UserNick.Description,
			JsonSchema:   factor.UserNick.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.TelNum.Factor,
			Name:         "手机号开关",
			Description:  factor.TelNum.Description,
			JsonSchema:   factor.TelNum.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.Single.Factor,
			Name:         "单一开关",
			Description:  factor.Single.Description,
			JsonSchema:   factor.Single.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.TimeRange.Factor,
			Name:         "时间范围开关",
			Description:  factor.TimeRange.Description,
			JsonSchema:   factor.TimeRange.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.UserId.Factor,
			Name:         "用户ID开关",
			Description:  factor.UserId.Description,
			JsonSchema:   factor.UserId.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.Location.Factor,
			Name:         "地理位置开关",
			Description:  factor.Location.Description,
			JsonSchema:   factor.Location.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.UserName.Factor,
			Name:         "用户名开关",
			Description:  factor.UserName.Description,
			JsonSchema:   factor.UserName.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
		{
			Factor:       factor.IP.Factor,
			Name:         "IP开关",
			Description:  factor.IP.Description,
			JsonSchema:   factor.IP.JsonSchema,
			NamespaceTag: nsTag,
			CommonModel: model.CommonModel{
				CreatedBy: createBy,
			},
		},
	}

	for _, meta := range metas {
		var existingMeta admin_model.SwitchFactor
		if tx.Where("factor = ? AND namespace_tag = ?", meta.Factor, meta.NamespaceTag).First(&existingMeta).Error != nil {
			tx.Create(meta)
		}
	}
}
