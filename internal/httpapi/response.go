package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Response struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      "OK",
		Message:   "success",
		Data:      data,
		RequestID: c.GetString("requestId"),
	})
}

func Fail(c *gin.Context, code string, message string) {
	status := httpStatusOf(code)
	c.JSON(status, Response{
		Code:      code,
		Message:   message,
		RequestID: c.GetString("requestId"),
	})
}

func FailWithStatus(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Response{
		Code:      code,
		Message:   message,
		RequestID: c.GetString("requestId"),
	})
}

func NewRequestID() string {
	return uuid.NewString()
}

func httpStatusOf(code string) int {
	switch code {
	case "OK":
		return http.StatusOK
	case "INVALID_PARAM", "INVALID_TIME_RANGE", "QUERY_RANGE_TOO_LARGE",
		"CONFIRMATION_REQUIRED", "INVALID_CONFIG_VALUE", "INVALID_STATUS",
		"STEP_NOT_COMPLETED", "INVALID_USERNAME", "INVALID_PASSWORD",
		"PASSWORD_MISMATCH", "WEAK_PASSWORD", "INVALID_JWT_SECRET",
		"DB_CONNECTION_FAILED", "DB_PATH_NOT_WRITABLE", "DB_CONNECTION_TIMEOUT",
		"STORAGE_PATH_INVALID", "INVALID_FILE_TYPE", "FILE_TOO_LARGE",
		"INVALID_FILENAME", "PLUGIN_ALREADY_EXISTS", "VERSION_ALREADY_EXISTS":
		return http.StatusBadRequest
	case "UNAUTHORIZED", "TOKEN_DISABLED":
		return http.StatusUnauthorized
	case "FORBIDDEN", "ALREADY_INITIALIZED":
		return http.StatusForbidden
	case "SW_NOT_FOUND", "PLUGIN_NOT_FOUND", "VERSION_NOT_FOUND",
		"CONFIG_NOT_FOUND", "PLUGIN_NOT_REVIEWABLE", "TOKEN_NOT_FOUND",
		"FILE_NOT_FOUND":
		return http.StatusNotFound
	case "RATE_LIMITED":
		return http.StatusTooManyRequests
	case "ACCOUNT_LOCKED":
		return http.StatusLocked
	case "REBUILD_IN_PROGRESS", "USER_ALREADY_EXISTS", "PROTECTED_USER":
		return http.StatusConflict
	case "INDEX_UNAVAILABLE":
		return http.StatusServiceUnavailable
	case "AUDIT_WRITE_FAILED", "INTERNAL_ERROR",
		"KEYGEN_FAILED", "CONFIG_WRITE_FAILED", "DB_INIT_FAILED",
		"SEED_FAILED", "MARKER_WRITE_FAILED", "RESTART_FAILED", "INIT_FAILED":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
