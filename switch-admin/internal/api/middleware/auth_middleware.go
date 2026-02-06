package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"gitee.com/fatzeng/switch-admin/info"
	"gitee.com/fatzeng/switch-admin/internal/dto"
	"gitee.com/fatzeng/switch-admin/internal/service"
	"gitee.com/fatzeng/switch-admin/internal/types"
	"gitee.com/fatzeng/switch-components/bc"

	"gitee.com/fatzeng/switch-admin/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var once = sync.Once{}
var userService *service.UserService

// AuthMiddleware jwt中间件做认证
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		respCtx := bc.NewRespContext(c)
		wrapper := bc.GetResp(respCtx)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoAuth("Authorization header is required")).ToResp())
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoAuth("Authorization header format must be Bearer {token}")).ToResp())
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoAuth("Invalid token")).ToResp())
			c.Abort()
			return
		}

		//解析用户关键信息
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userID", claims["sub"])
			c.Set("userName", claims["username"])
			c.Set("exp", claims["exp"])
			//不选择没有
			c.Set("selectNamespace", claims["selectNamespace"])

			once.Do(func() {
				userService = service.NewUserService()
			})

			userInfo, err := userService.GetUserInfo(info.GetUserId(c))
			if err != nil {
				c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoAuth("GetUserInfo fail invalid token")).ToResp())
				c.Abort()
				return
			}
			c.Set("userInfo", &dto.UserInfo{
				User:            userInfo,
				SelectNamespace: info.GetSelectNamespace(c),
			})
		} else {
			c.JSON(http.StatusForbidden, wrapper.SetErrors(types.NoAuth("Invalid token")).ToResp())
			c.Abort()
			return
		}

		c.Next()
	}
}
