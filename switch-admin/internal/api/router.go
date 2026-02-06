package api

import (
	"gitee.com/fatzeng/switch-admin/internal/api/controller"
	"gitee.com/fatzeng/switch-admin/internal/api/middleware"
	"gitee.com/fatzeng/switch-admin/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 使用CORS中间件
	router.Use(cors.Default())

	// 登录注册
	authController := controller.NewAuthController(cfg)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}

	authRefresh := router.Group("/auth")
	authRefresh.Use(middleware.AuthMiddleware(cfg))
	{
		authRefresh.POST("/refresh-token", authController.RefreshToken)
	}

	// 定义开关相关的核心api
	apiV1 := router.Group("/api/v1")
	apiV2 := router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(cfg))
	{
		namespaceController := controller.NewNamespaceController()
		envController := controller.NewEnvironmentController()
		switchController := controller.NewSwitchController()
		approvalController := controller.NewApprovalController()
		userController := controller.NewUserController()
		permissionController := controller.NewPermissionController()

		// 获取当前用户的当前信息 包括角色权限
		apiV1.GET("/user", userController.GetMe)
		apiV2.POST("/user-like", userController.GetAllUserLike)
		// 查询用户列表（用户管理页面）
		apiV1.POST("/users", middleware.RBACMiddleware("users:watch"), userController.ListUsers)
		// 查询用户列表（用于开关创建/编辑时选择审批人）
		apiV1.POST("/users-permission", middleware.RBACMiddleware("users:list-for-switch"), userController.ListUsersPermissions)

		// 给用户分配角色
		apiV1.POST("/assignRoles", middleware.RBACMiddleware("users:assign-roles"), permissionController.AssignRoles)
		// 查询空间内的角色/权限
		apiV1.POST("/roles-permission", middleware.RBACMiddleware("roles:watch"), permissionController.RolesPermission)
		// 新增角色
		apiV1.POST("/roles-create", middleware.RBACMiddleware("roles:create"), permissionController.InsertRole)
		// 修改角色
		apiV1.POST("/roles-edit", middleware.RBACMiddleware("roles:edit"), permissionController.UpdateRole)

		// 查询空间内的权限
		apiV1.POST("/permissions", middleware.RBACMiddleware("permissions:watch"), permissionController.Permissions)
		// 新增权限
		apiV1.POST("/permissions-create", middleware.RBACMiddleware("permissions:create"), permissionController.InsertPermission)
		// 修改权限
		apiV1.POST("/permissions-edit", middleware.RBACMiddleware("permissions:edit"), permissionController.UpdatePermission)

		// 获取所有的命名空间（仅超级管理员）
		apiV1.POST("/namespaces", middleware.RBACMiddleware("canSuperAdmin"), namespaceController.GetAllNamespaces)
		// 获取所有的命名空间 用作下拉列表
		apiV1.POST("/namespaces-like", namespaceController.GetAllNamespacesLike)
		// 创建命名空间（无需权限认证）
		apiV1.POST("/namespace-create", namespaceController.CreateNamespace)
		//编辑命名空间(只能编辑名称跟描述)
		apiV1.POST("/namespace-edit", middleware.RBACMiddleware("namespaces:edit"), namespaceController.UpdateNamespace)

		//获取某个命名空间下的所有环境
		apiV1.GET("/namespaces/:namespaceTag/environments", middleware.RBACMiddleware("namespace-envs:watch"), envController.GetEnvironmentsByNamespace)
		//获取环境信息
		apiV1.POST("/envs", middleware.RBACMiddleware("envs:watch"), envController.GetAllEnvs)
		//创建环境
		apiV1.POST("/env-create", middleware.RBACMiddleware("envs:create"), envController.CreateEnv)
		//编辑环境
		apiV1.POST("/env-edit", middleware.RBACMiddleware("envs:edit"), envController.UpdateEnv)
		//发布环境
		apiV1.POST("/envs-publish", middleware.RBACMiddleware("envs:publish"), envController.PublishEnv)

		//开关相关的操作
		//创建开关以及开关配置以及审批信息(开关配置对照不同环境将生成多份)(审批信息对照着不同的环境)
		apiV1.POST("/switches", middleware.RBACMiddleware("switches:watch"), switchController.SwitchList)
		//获取开关详情以及各个环境下的开关配置 开关详情只有推送到不同环境。不同环境的开关才生效
		apiV1.GET("/switch/:id", middleware.RBACMiddleware("switches:watch"), switchController.GetSwitchDetails)
		//创建开关
		apiV1.POST("/switch-create", middleware.RBACMiddleware("switches:create"), switchController.CreateSwitch)
		//开关修改(同时更新当前节点为env[0])
		apiV1.POST("/switch-edit", middleware.RBACMiddleware("switches:edit"), switchController.UpdateSwitch)
		//推送开关
		apiV1.POST("/switch-submit-push", middleware.RBACMiddleware("switches:push"), switchController.SubmitSwitchPush)
		//创建因子信息
		apiV1.POST("/switch-factor-create", middleware.RBACMiddleware("switch-factors:create"), switchController.CreateSwitchFactor)
		//编辑因子信息
		apiV1.POST("/switch-factor-edit", middleware.RBACMiddleware("switch-factors:edit"), switchController.UpdateSwitchFactor)
		//因子信息列表
		apiV1.POST("/switch-factors", middleware.RBACMiddleware("switch-factors:watch"), switchController.MetaSwitchList)
		//模糊获取空间下所有的因子信息列表
		apiV1.POST("/switch-factors-like", middleware.RBACMiddleware("switch-factors:watch"), switchController.MetaSwitchLike)

		// 开关审批(受邀人,发起人)
		approvalRoutes := apiV1.Group("/approvals")
		{
			//条件查询 受邀人 发起人的单子
			approvalRoutes.POST("/my-requests", middleware.RBACMiddleware("approvals:watch"), approvalController.GetMyRequests)
			//非条件查询 不区分受邀人 发起人的单子(超级管理员)
			approvalRoutes.POST("/all-requests", middleware.RBACMiddleware("approvals:watch"), approvalController.GetAllRequests)
			//审批一个请求
			approvalRoutes.POST("/:id/approve", middleware.RBACMiddleware("approvals:approve"), approvalController.JudgeRequest)
			//申请加入空间
			approvalRoutes.POST("/applyToJoin", approvalController.ApplyToJoin)
		}
	}

	return router
}
