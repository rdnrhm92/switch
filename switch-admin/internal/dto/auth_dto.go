package dto

import "gitee.com/fatzeng/switch-admin/internal/admin_model"

// RegisterReq 定义一个注册的请求体
type RegisterReq struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	AutoLogin bool   `json:"autoLogin"`
}

type RegisterRes struct {
	*admin_model.User
	Token string `json:"token"`
}

// LoginReq 定义一个登录的请求体
type LoginReq struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	AutoLogin bool   `json:"autoLogin"`
}

// RefreshToken 刷新令牌
type RefreshToken struct {
	UserId          uint   `json:"userId"`
	SelectNamespace string `json:"selectNamespace"`
}
