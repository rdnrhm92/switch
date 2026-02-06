package service

import (
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/admin_model"
	"gitee.com/fatzeng/switch-admin/internal/config"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthService 提供认证相关的服务。
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtConfig *config.JWTConfig
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:  repository.NewUserRepository(),
		jwtConfig: cfg.JWT,
	}
}

// Register 创建新用户
func (s *AuthService) Register(ctx *gin.Context, req *dto.RegisterReq) (*dto.RegisterRes, error) {
	// 检查用户名是否已存在
	_, err := s.userRepo.FindByUsername(req.Username)
	if err == nil {
		return nil, fmt.Errorf("username '%s' already exists", req.Username)
	}
	isFalse := false
	user := &admin_model.User{
		Username:     req.Username,
		IsSuperAdmin: &isFalse,
	}

	if err = user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err = s.userRepo.Create(user); err != nil {
		return nil, err
	}

	var exp int64 = 0
	if req.AutoLogin {
		exp = time.Now().Add(time.Hour * time.Duration(s.jwtConfig.AutoLoginExpirationHours)).Unix()
	} else {
		exp = time.Now().Add(time.Hour * time.Duration(s.jwtConfig.ExpirationHours)).Unix()
	}

	// 创建 JWT 令牌
	token, err := s.createToken(ctx, user.ID, req.Username, exp, "")
	if err != nil {
		return nil, err
	}

	return &dto.RegisterRes{
		User:  user,
		Token: token,
	}, nil
}

// Login 验证凭据，如果有效则返回 JWT 令牌
func (s *AuthService) Login(ctx *gin.Context, username string, password string, autoLogin bool) (string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", fmt.Errorf("invalid username or password")
	}

	if !user.CheckPassword(password) {
		return "", fmt.Errorf("invalid username or password")
	}

	var exp int64 = 0
	if autoLogin {
		exp = time.Now().Add(time.Hour * time.Duration(s.jwtConfig.AutoLoginExpirationHours)).Unix()
	} else {
		exp = time.Now().Add(time.Hour * time.Duration(s.jwtConfig.ExpirationHours)).Unix()
	}

	// 创建 JWT 令牌
	return s.createToken(ctx, user.ID, username, exp, "")
}

func (s *AuthService) createToken(ctx *gin.Context, userId uint, userName string, exp int64, selectNamespace string) (string, error) {
	// 创建 JWT 令牌
	claims := jwt.MapClaims{
		"sub":      userId,
		"username": userName,
		"exp":      exp,
		"iat":      time.Now().Unix(),
	}
	if selectNamespace != "" {
		claims["selectNamespace"] = selectNamespace
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return t, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx *gin.Context, refreshToken *dto.RefreshToken) (string, error) {
	user, err := s.userRepo.FindByUserId(refreshToken.UserId)
	if err != nil {
		return "", fmt.Errorf("can not find user[%v]: %w", refreshToken.UserId, err)
	}

	exp := info.GetExp(ctx)

	return s.createToken(ctx, user.ID, user.Username, exp, refreshToken.SelectNamespace)
}
