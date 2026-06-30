package middleware

import (
	"net/http"
	"strings"

	"signature-menu-backend/internal/httpx"
	"signature-menu-backend/pkg/token"

	"github.com/gin-gonic/gin"
)

const userIDKey = "userID"

func Auth(tokens *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "请先登录")
			c.Abort()
			return
		}

		userID, err := tokens.Validate(strings.TrimSpace(authHeader[7:]))
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "登录状态已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	value, ok := c.Get(userIDKey)
	if !ok {
		return ""
	}
	userID, _ := value.(string)
	return userID
}
