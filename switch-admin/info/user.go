package info

import (
	"context"

	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-sdk-core/logger"
	"gitee.com/fatzeng/switch-sdk-core/tool"
)

func GetUserId(ctx context.Context) uint {
	userID := ctx.Value("userID")
	intUserId, err := tool.ToInt(userID)
	if err != nil {
		logger.Logger.Errorf("get user id to Int err: %v", err)
		return 0
	}
	return uint(intUserId)
}

func GetExp(ctx context.Context) int64 {
	exp := ctx.Value("exp")
	expVar, err := tool.ToInt64(exp)
	if err != nil {
		logger.Logger.Errorf("get exp to Int64 err: %v", err)
		return 0
	}
	return expVar
}

func GetSelectNamespace(ctx context.Context) string {
	nsTag := ctx.Value("selectNamespace")
	if nsTag == nil {
		return ""
	}
	return nsTag.(string)
}

func GetUserName(ctx context.Context) string {
	userName := ctx.Value("userName")
	return tool.ToString(userName)
}

func GetUserInfo(ctx context.Context) *dto.UserInfo {
	userInfo := ctx.Value("userInfo")
	user, ok := userInfo.(*dto.UserInfo)
	if !ok {
		logger.Logger.Errorf("get userinfo err type not *admin_model.User")
		return nil
	}
	return user
}
