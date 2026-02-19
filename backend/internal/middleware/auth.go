package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
)

const UserIDKey = "userID"

func Auth(tokenMgr *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := extractBearer(c)
		if t == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": apperror.ErrUnauthorized.Error()})
			return
		}
		claims, err := tokenMgr.ParseAccessToken(t)
		if err != nil {
			status := apperror.HTTPStatus(err)
			c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func OptionalAuth(tokenMgr *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := extractBearer(c)
		if t != "" {
			if claims, err := tokenMgr.ParseAccessToken(t); err == nil {
				c.Set(UserIDKey, claims.UserID)
			}
		}
		c.Next()
	}
}

func extractBearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

func MustUserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(UserIDKey)
	id, _ := v.(uuid.UUID)
	return id
}
