package repository

import "gitee.com/fatzeng/switch-admin/internal/admin_model"

// SuperAdmin 超级管理员 拥有所有命名空间的所有权限
var SuperAdmin = &admin_model.User{}

// AdminRole 管理员角色，拥有命名空间下的所有权限
var AdminRole = &admin_model.Role{}

// DeveloperRole 开发者角色，可以管理开关、审批，以及查看相关资源
var DeveloperRole = &admin_model.Role{}

// ObserverRole 观察者角色，拥有所有资源的查看权限，但无操作权限
var ObserverRole = &admin_model.Role{}

// OrdinaryRole 普通用户角色，拥有基础的查看权限
var OrdinaryRole = &admin_model.Role{}
