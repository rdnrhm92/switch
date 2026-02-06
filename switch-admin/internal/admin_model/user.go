package admin_model

import (
	"gitee.com/fatzeng/switch-sdk-core/model"
	"golang.org/x/crypto/bcrypt"
)

// User 用户表
type User struct {
	model.CommonModel
	Username     string `gorm:"size:50;not null;unique;comment:用户名" json:"username"`
	Password     string `gorm:"size:255;not null;comment:哈希后的密码" json:"-"`
	IsSuperAdmin *bool  `gorm:"not null;comment:是否为超级管理员;default:1" json:"is_super_admin"` //超级管理员
	//表关联数据
	NamespaceMembers []*NamespaceMembers `gorm:"foreignKey:UserId" json:"namespaceMembers,omitempty"`
}

func (*User) TableName() string {
	return "users"
}

// SetPassword 密码hash
func (u *User) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

// CheckPassword 密码校验
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
