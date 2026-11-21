package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pclsguserver/internal/service/auth"
)

func fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"code": code, "message": message, "requestId": c.GetString("requestId")})
	c.Abort()
}

const (
	PermSoftwarePublish = "software.publish"
	PermPluginSubmit    = "plugin.submit"
	PermPluginManage    = "plugin.manage"
	PermPluginReview    = "plugin.review"
	PermIndexRebuild    = "index.rebuild"
	PermIndexView       = "index.view"
	PermMetricsView     = "metrics.view"
	PermAuditView       = "audit.view"
	PermUserManage      = "user.manage"
	PermConfigManage    = "config.manage"
)

func AuthMiddleware(authSvc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
			c.Abort()
			return
		}
		claims, err := authSvc.ParseToken(parts[1])
		if err != nil {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			c.Abort()
			return
		}
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRoles", claims.Roles)
		c.Set("token", parts[1])
		c.Next()
	}
}

type PermissionCache struct {
	mu    sync.RWMutex
	cache map[string][]string
	db    *gorm.DB
}

func NewPermissionCache(db *gorm.DB) *PermissionCache {
	return &PermissionCache{cache: make(map[string][]string), db: db}
}

func (pc *PermissionCache) Get(userID string) ([]string, error) {
	pc.mu.RLock()
	if perms, ok := pc.cache[userID]; ok {
		pc.mu.RUnlock()
		return perms, nil
	}
	pc.mu.RUnlock()

	var roles []struct {
		Permissions []string `gorm:"column:permissions;serializer:json"`
	}
	err := pc.db.Table("roles").
		Joins("JOIN user_roles ON user_roles.roleId = roles.id").
		Where("user_roles.userId = ?", userID).
		Scan(&roles).Error
	if err != nil {
		return nil, err
	}
	permSet := make(map[string]bool)
	for _, r := range roles {
		for _, p := range r.Permissions {
			permSet[p] = true
		}
	}
	var perms []string
	for p := range permSet {
		perms = append(perms, p)
	}
	pc.mu.Lock()
	pc.cache[userID] = perms
	pc.mu.Unlock()
	return perms, nil
}

func (pc *PermissionCache) Invalidate(userID string) {
	pc.mu.Lock()
	delete(pc.cache, userID)
	pc.mu.Unlock()
}

func (pc *PermissionCache) InvalidateAll() {
	pc.mu.Lock()
	pc.cache = make(map[string][]string)
	pc.mu.Unlock()
}

func RBACMiddleware(requiredPermission string, permCache *PermissionCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userId")
		if userID == "" {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
			c.Abort()
			return
		}
		perms, err := permCache.Get(userID)
		if err != nil {
			fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load permissions")
			c.Abort()
			return
		}
		for _, p := range perms {
			if p == "*" || p == requiredPermission {
				c.Next()
				return
			}
		}
		fail(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		c.Abort()
	}
}
