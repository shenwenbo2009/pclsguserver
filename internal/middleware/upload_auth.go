package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pclsguserver/internal/service/upload"
)

func UploadTokenAuthMiddleware(tokenSvc *upload.UploadTokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Upload-Token")
		if token == "" {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "缺少上传 Token")
			return
		}

		uploadToken, err := tokenSvc.Validate(token)
		if err != nil {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token 无效")
			return
		}

		if uploadToken.Status == "disabled" {
			fail(c, http.StatusUnauthorized, "TOKEN_DISABLED", "Token 已禁用")
			return
		}

		path := c.Request.URL.Path
		var requiredPerm string
		switch {
		case strings.Contains(path, "/plugin"):
			requiredPerm = "plugin"
		case strings.Contains(path, "/software"):
			requiredPerm = "software"
		case strings.Contains(path, "/file"):
			requiredPerm = "file"
		}

		if requiredPerm != "" {
			hasPerm := false
			for _, p := range uploadToken.Permissions {
				if p == requiredPerm || p == "*" {
					hasPerm = true
					break
				}
			}
			if !hasPerm {
				fail(c, http.StatusForbidden, "FORBIDDEN", "Token 权限不足")
				return
			}
		}

		c.Set("uploaderIdentity", uploadToken.UploaderIdentity)
		c.Set("uploadTokenId", uploadToken.ID)
		c.Next()
	}
}
