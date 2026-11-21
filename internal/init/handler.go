package init

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type InitHandler struct {
	executor   *InitExecutor
	configPath string
}

func NewInitHandler(configPath string) *InitHandler {
	return &InitHandler{
		executor:   NewInitExecutor(configPath),
		configPath: configPath,
	}
}

func (h *InitHandler) RegisterInitRoutes(r *gin.Engine) {
	api := r.Group("/api/init")
	api.GET("/status", h.handleStatus)
	api.POST("/validate", h.handleValidate)
	api.POST("/generate-jwt", h.handleGenerateJwt)
	api.POST("/test-db", h.handleTestDb)
	api.POST("/test-storage", h.handleTestStorage)
	api.POST("/complete", h.handleComplete)
}

func (h *InitHandler) handleStatus(c *gin.Context) {
	isFirstRun, status := GetInitStatus(h.configPath)
	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "success",
		"data": gin.H{
			"isFirstRun": isFirstRun,
			"initStatus": status,
		},
	})
}

func (h *InitHandler) handleValidate(c *gin.Context) {
	var body struct {
		Step int                    `json:"step"`
		Data map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PARAM", "message": "invalid body"})
		return
	}
	valid, errs := ValidateStep(body.Step, body.Data)
	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "success",
		"data": gin.H{
			"valid":  valid,
			"errors": errs,
		},
	})
}

func (h *InitHandler) handleGenerateJwt(c *gin.Context) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "KEYGEN_FAILED", "message": "密钥自动生成失败，请手动输入"})
		return
	}
	secret := base64.StdEncoding.EncodeToString(key)
	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "success",
		"data":    gin.H{"jwtSecret": secret},
	})
}

func (h *InitHandler) handleTestDb(c *gin.Context) {
	var body struct {
		DbDriver   string `json:"dbDriver"`
		SqlitePath string `json:"sqlitePath"`
		DbHost     string `json:"dbHost"`
		DbPort     string `json:"dbPort"`
		DbName     string `json:"dbName"`
		DbUser     string `json:"dbUser"`
		DbPassword string `json:"dbPassword"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PARAM", "message": "invalid body"})
		return
	}

	if body.DbDriver == "sqlite" {
		path := body.SqlitePath
		if path == "" {
			path = "./data/meta.db"
		}
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "DB_PATH_NOT_WRITABLE", "message": "路径不可写"})
			return
		}
		db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "DB_CONNECTION_FAILED", "message": err.Error()})
			return
		}
		sqlDB, _ := db.DB()
		sqlDB.Close()
		c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "success", "data": gin.H{"success": true}})
		return
	}

	if body.DbDriver == "postgres" {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			body.DbHost, body.DbPort, body.DbUser, body.DbPassword, body.DbName)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "DB_CONNECTION_FAILED", "message": err.Error()})
			return
		}
		sqlDB, _ := db.DB()
		sqlDB.Close()
		c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "success", "data": gin.H{"success": true}})
		return
	}

	if body.DbDriver == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			body.DbUser, body.DbPassword, body.DbHost, body.DbPort, body.DbName)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "DB_CONNECTION_FAILED", "message": err.Error()})
			return
		}
		sqlDB, _ := db.DB()
		sqlDB.Close()
		c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "success", "data": gin.H{"success": true}})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PARAM", "message": "unsupported db driver"})
}

func (h *InitHandler) handleTestStorage(c *gin.Context) {
	var body struct {
		StorageDriver string `json:"storageDriver"`
		StoragePath   string `json:"storagePath"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PARAM", "message": "invalid body"})
		return
	}

	if body.StorageDriver == "fs" {
		path := body.StoragePath
		if path == "" {
			path = "./data"
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "STORAGE_PATH_INVALID", "message": "存储路径不可写"})
			return
		}
		testFile := filepath.Join(path, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "STORAGE_PATH_INVALID", "message": "存储路径不可写"})
			return
		}
		os.Remove(testFile)
		c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "success", "data": gin.H{"success": true}})
		return
	}

	if body.StorageDriver == "s3" {
		c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "success", "data": gin.H{"success": true}})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PARAM", "message": "unsupported storage driver"})
}

func (h *InitHandler) handleComplete(c *gin.Context) {
	if !IsFirstRun(h.configPath) {
		c.JSON(http.StatusForbidden, gin.H{"code": "ALREADY_INITIALIZED", "message": "系统已初始化"})
		return
	}

	var cfg InitConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PARAM", "message": "invalid body"})
		return
	}

	if err := h.executor.Execute(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INIT_FAILED", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "success",
		"data": gin.H{
			"success": true,
			"message": "初始化完成，服务正在重启...",
		},
	})
}
