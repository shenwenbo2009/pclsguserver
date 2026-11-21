package httpapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pclsguserver/internal/apidoc"
	"pclsguserver/internal/audit"
	"pclsguserver/internal/config"
	"pclsguserver/internal/index"
	"pclsguserver/internal/metrics"
	"pclsguserver/internal/middleware"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/service/auth"
	configsvc "pclsguserver/internal/service/config"
	indexsvc "pclsguserver/internal/service/index"
	"pclsguserver/internal/service/plugin"
	"pclsguserver/internal/service/review"
	"pclsguserver/internal/service/update"
	uploadsvc "pclsguserver/internal/service/upload"
	"pclsguserver/internal/service/user"
	"pclsguserver/internal/store"
)

type App struct {
	cfg               *config.Config
	db                *gorm.DB
	binaryStore       store.BinaryStore
	swRepo            *repo.SoftwareVersionRepo
	pluginRepo        *repo.PluginRepo
	reviewRepo        *repo.ReviewRecordRepo
	userRepo          *repo.AdminUserRepo
	roleRepo          *repo.RoleRepo
	auditRepo         *repo.AuditRepo
	metricsRepo       *repo.MetricsRepo
	configRepo        *repo.ConfigRepo
	uploadTokenRepo   *repo.UploadTokenRepo
	fileRecordRepo    *repo.FileRecordRepo
	authSvc           *auth.AuthService
	auditSvc          *audit.Service
	updateSvc         *update.Service
	pluginSvc         *plugin.Service
	reviewSvc         *review.Service
	indexSvc          *indexsvc.Service
	metricsSvc        *metrics.Service
	userSvc           *user.Service
	configSvc         *configsvc.Service
	uploadTokenSvc    *uploadsvc.UploadTokenService
	pluginUploadSvc   *uploadsvc.PluginUploadService
	softwareUploadSvc *uploadsvc.SoftwareUploadService
	fileUploadSvc     *uploadsvc.FileUploadService
	uploadRateLimiter *middleware.UploadRateLimiter
	fileValidator     *uploadsvc.FileValidator
	indexEngine       *index.Engine
	indexBuilder      *index.Builder
	permCache         *middleware.PermissionCache
}

func NewApp(cfg *config.Config, db *gorm.DB, bs store.BinaryStore) *App {
	swRepo := repo.NewSoftwareVersionRepo(db)
	pluginRepo := repo.NewPluginRepo(db)
	reviewRepo := repo.NewReviewRecordRepo(db)
	userRepo := repo.NewAdminUserRepo(db)
	roleRepo := repo.NewRoleRepo(db)
	auditRepo := repo.NewAuditRepo(db)
	metricsRepo := repo.NewMetricsRepo(db)
	configRepo := repo.NewConfigRepo(db)
	uploadTokenRepo := repo.NewUploadTokenRepo(db)
	fileRecordRepo := repo.NewFileRecordRepo(db)

	auditSvc := audit.NewService(auditRepo)
	authSvc := auth.NewAuthService(userRepo, roleRepo, auditRepo, cfg)
	metricsSvc := metrics.NewService(metricsRepo)
	indexEngine := index.NewEngine()
	indexBuilder := index.NewBuilder(indexEngine, pluginRepo, swRepo, cfg.Server.DataDir, cfg.Index.FormatVersion)
	updateSvc := update.NewService(swRepo, bs, metricsSvc, auditSvc, db)
	pluginSvc := plugin.NewService(pluginRepo, bs, metricsSvc, auditSvc, indexEngine, indexBuilder, db, cfg.Index.DegradePolicy)
	reviewSvc := review.NewService(pluginRepo, reviewRepo, auditSvc, indexBuilder, db)
	indexSvc2 := indexsvc.NewService(indexBuilder, auditSvc)
	permCache := middleware.NewPermissionCache(db)
	userSvc := user.NewService(userRepo, roleRepo, auditSvc, permCache, db)
	configSvc := configsvc.NewService(configRepo, auditSvc)

	fileValidator := uploadsvc.NewFileValidator(cfg.Upload.PluginAllowedTypes, cfg.Upload.FileAllowedTypes)
	uploadTokenSvc := uploadsvc.NewUploadTokenService(uploadTokenRepo)
	pluginUploadSvc := uploadsvc.NewPluginUploadService(fileValidator, bs, pluginRepo, auditSvc, db, cfg.Upload.PluginMaxSize)
	softwareUploadSvc := uploadsvc.NewSoftwareUploadService(fileValidator, bs, swRepo, auditSvc, db, cfg.Upload.SoftwareMaxSize)
	fileUploadSvc := uploadsvc.NewFileUploadService(fileValidator, bs, fileRecordRepo, auditSvc, db, cfg.Upload.FileMaxSize)
	uploadRateLimiter := middleware.NewUploadRateLimiter()

	return &App{
		cfg: cfg, db: db, binaryStore: bs,
		swRepo: swRepo, pluginRepo: pluginRepo, reviewRepo: reviewRepo,
		userRepo: userRepo, roleRepo: roleRepo, auditRepo: auditRepo,
		metricsRepo: metricsRepo, configRepo: configRepo,
		uploadTokenRepo: uploadTokenRepo, fileRecordRepo: fileRecordRepo,
		authSvc: authSvc, auditSvc: auditSvc, updateSvc: updateSvc,
		pluginSvc: pluginSvc, reviewSvc: reviewSvc, indexSvc: indexSvc2,
		metricsSvc: metricsSvc, userSvc: userSvc, configSvc: configSvc,
		uploadTokenSvc: uploadTokenSvc, pluginUploadSvc: pluginUploadSvc,
		softwareUploadSvc: softwareUploadSvc, fileUploadSvc: fileUploadSvc,
		uploadRateLimiter: uploadRateLimiter, fileValidator: fileValidator,
		indexEngine: indexEngine, indexBuilder: indexBuilder, permCache: permCache,
	}
}

func (a *App) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	a.registerUpdateRoutes(api)
	a.registerPluginRoutes(api)
	a.registerUploadRoutes(api)
	a.registerAdminRoutes(api)
	r.GET("/health", a.health)
	r.GET("/index.json", a.serveIndexJson)
	r.GET("/static/index.json", a.serveIndexJson)
	r.GET("/files/*path", a.serveFile)
}

func (a *App) health(c *gin.Context) {
	dbOK := a.db != nil
	idxOK := a.indexEngine.IsReady()
	status := "ok"
	if !dbOK {
		status = "degraded"
	}
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"deps": gin.H{
			"db":      dbOK,
			"index":   idxOK,
			"storage": true,
		},
	})
}

func (a *App) serveIndexJson(c *gin.Context) {
	data, err := a.indexBuilder.LoadIndexJson()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "index.json not available")
		return
	}
	c.Header("Cache-Control", "public, max-age=5")
	c.Data(http.StatusOK, "application/json", data)
}

func (a *App) registerUpdateRoutes(g *gin.RouterGroup) {
	g.GET("/update/check", a.updateCheck)
	g.GET("/update/download", a.updateDownload)
	g.POST("/update/report", a.updateReport)
}

func (a *App) updateCheck(c *gin.Context) {
	swCode := c.Query("swCode")
	currentVersion := c.Query("currentVersion")
	if swCode == "" || currentVersion == "" {
		Fail(c, "INVALID_PARAM", "swCode and currentVersion required")
		return
	}
	result, err := a.updateSvc.Check(c.Request.Context(), swCode, currentVersion, c.ClientIP())
	if err != nil {
		Fail(c, "SW_NOT_FOUND", "software not found")
		return
	}
	OK(c, result)
}

func (a *App) updateDownload(c *gin.Context) {
	swCode := c.Query("swCode")
	version := c.Query("version")
	if swCode == "" || version == "" {
		Fail(c, "INVALID_PARAM", "swCode and version required")
		return
	}
	v, err := a.updateSvc.GetForDownload(c.Request.Context(), swCode, version, c.ClientIP())
	if err != nil {
		Fail(c, "VERSION_NOT_FOUND", "version not found")
		return
	}
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		offset, length, ok := parseRange(rangeHeader, v.Size)
		if ok {
			reader, err := a.updateSvc.ReadPackageRange(c.Request.Context(), swCode, version, offset, length)
			if err != nil {
				Fail(c, "INTERNAL_ERROR", "read failed")
				return
			}
			defer reader.Close()
			c.Header("Content-Range", "bytes "+strconv.FormatInt(offset, 10)+"-"+strconv.FormatInt(offset+length-1, 10)+"/"+strconv.FormatInt(v.Size, 10))
			c.DataFromReader(http.StatusPartialContent, length, "application/octet-stream", reader, nil)
			a.updateSvc.IncrementDownloadCount(swCode, version, c.ClientIP())
			return
		}
	}
	reader, err := a.updateSvc.ReadPackage(c.Request.Context(), swCode, version)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "read failed")
		return
	}
	defer reader.Close()
	c.Header("X-Checksum", v.Checksum)
	c.DataFromReader(http.StatusOK, v.Size, "application/octet-stream", reader, nil)
	a.updateSvc.IncrementDownloadCount(swCode, version, c.ClientIP())
}

func (a *App) updateReport(c *gin.Context) {
	var body struct {
		SwCode    string `json:"swCode"`
		Version   string `json:"version"`
		EventType string `json:"eventType"`
		ClientID  string `json:"clientId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	if body.EventType == "query" {
		a.metricsSvc.RecordEvent(model.MetricsDimensionSoftwareActivity, body.SwCode+":"+body.Version, body.ClientID)
	} else if body.EventType == "download" {
		a.metricsSvc.RecordEvent(model.MetricsDimensionSoftwareDownload, body.SwCode+":"+body.Version, body.ClientID)
	}
	OK(c, nil)
}

func (a *App) registerPluginRoutes(g *gin.RouterGroup) {
	g.GET("/plugins", a.pluginList)
	g.GET("/plugins/search", a.pluginSearch)
	g.GET("/plugins/:id", a.pluginDetail)
	g.GET("/plugins/:id/download", a.pluginDownload)
}

func (a *App) pluginList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	category := c.Query("category")
	ps, total, err := a.pluginSvc.ListPublished(page, size, category)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": ps, "total": total, "page": page, "size": size})
}

func (a *App) pluginSearch(c *gin.Context) {
	keyword := c.Query("keyword")
	category := c.Query("category")
	name := c.Query("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := a.pluginSvc.Search(keyword, category, name, page, size)
	if err != nil {
		if strings.Contains(err.Error(), "index unavailable") {
			Fail(c, "INDEX_UNAVAILABLE", "index not ready")
			return
		}
		Fail(c, "INTERNAL_ERROR", "search failed")
		return
	}
	OK(c, gin.H{"items": result.Items, "total": result.Total, "page": page, "size": size})
}

func (a *App) pluginDetail(c *gin.Context) {
	id := c.Param("id")
	p, err := a.pluginSvc.GetPublished(id)
	if err != nil {
		Fail(c, "PLUGIN_NOT_FOUND", "plugin not found")
		return
	}
	OK(c, p)
}

func (a *App) pluginDownload(c *gin.Context) {
	id := c.Param("id")
	version := c.Query("version")
	p, err := a.pluginSvc.GetForDownload(c.Request.Context(), id, version, c.ClientIP())
	if err != nil {
		Fail(c, "PLUGIN_NOT_FOUND", "plugin not found")
		return
	}
	reader, err := a.pluginSvc.ReadPackage(c.Request.Context(), id, p.Version)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "read failed")
		return
	}
	defer reader.Close()
	c.Header("X-Checksum", p.Checksum)
	c.DataFromReader(http.StatusOK, p.Size, "application/octet-stream", reader, nil)
	a.pluginSvc.IncrementDownloadCount(id, c.ClientIP())
}

func (a *App) registerAdminRoutes(g *gin.RouterGroup) {
	admin := g.Group("/admin")
	admin.POST("/login", a.adminLogin)
	admin.POST("/auth/refresh", a.adminRefresh)
	admin.POST("/logout", a.adminLogout)

	authed := admin.Group("", middleware.AuthMiddleware(a.authSvc))
	authed.GET("/auth/permissions", a.adminPermissions)
	authed.GET("/audit", middleware.RBACMiddleware(middleware.PermAuditView, a.permCache), a.adminAuditList)
	authed.GET("/metrics", middleware.RBACMiddleware(middleware.PermMetricsView, a.permCache), a.adminMetrics)

	sw := authed.Group("/software/versions", middleware.RBACMiddleware(middleware.PermSoftwarePublish, a.permCache))
	sw.GET("", a.adminSwList)
	sw.POST("", a.adminSwPublish)
	sw.PUT("/:id/offline", a.adminSwOffline)
	sw.DELETE("/:id", a.adminSwDelete)

	pl := authed.Group("/plugins")
	pl.GET("", a.adminPluginList)
	pl.POST("", middleware.RBACMiddleware(middleware.PermPluginSubmit, a.permCache), a.adminPluginSubmit)
	pl.PUT("/:id/offline", middleware.RBACMiddleware(middleware.PermPluginManage, a.permCache), a.adminPluginOffline)
	pl.DELETE("/:id", middleware.RBACMiddleware(middleware.PermPluginManage, a.permCache), a.adminPluginDelete)
	pl.POST("/:id/resubmit", middleware.RBACMiddleware(middleware.PermPluginSubmit, a.permCache), a.adminPluginResubmit)
	pl.GET("/reviewing", middleware.RBACMiddleware(middleware.PermPluginReview, a.permCache), a.adminPluginReviewing)
	pl.POST("/:id/review", middleware.RBACMiddleware(middleware.PermPluginReview, a.permCache), a.adminPluginReview)

	idx := authed.Group("/index")
	idx.POST("/rebuild", middleware.RBACMiddleware(middleware.PermIndexRebuild, a.permCache), a.adminIndexRebuild)
	idx.GET("/status", middleware.RBACMiddleware(middleware.PermIndexView, a.permCache), a.adminIndexStatus)

	usr := authed.Group("/users", middleware.RBACMiddleware(middleware.PermUserManage, a.permCache))
	usr.GET("", a.adminUserList)
	usr.POST("", a.adminUserCreate)
	usr.GET("/:id", a.adminUserGet)
	usr.PUT("/:id/disable", a.adminUserDisable)
	usr.PUT("/:id/enable", a.adminUserEnable)
	usr.DELETE("/:id", a.adminUserDelete)
	usr.PUT("/:id/roles", a.adminUserAssignRoles)

	roles := authed.Group("/roles", middleware.RBACMiddleware(middleware.PermUserManage, a.permCache))
	roles.GET("", a.adminRoleList)
	roles.POST("", a.adminRoleCreate)
	roles.PUT("/:id", a.adminRoleUpdate)
	roles.DELETE("/:id", a.adminRoleDelete)

	cfg := authed.Group("/config", middleware.RBACMiddleware(middleware.PermConfigManage, a.permCache))
	cfg.GET("", a.adminConfigList)
	cfg.PUT("/:key", a.adminConfigUpdate)

	authed.POST("/docs/regenerate", middleware.RBACMiddleware(middleware.PermConfigManage, a.permCache), a.adminDocsRegenerate)

	tokens := authed.Group("/upload-tokens", middleware.RBACMiddleware(middleware.PermConfigManage, a.permCache))
	tokens.GET("", a.adminUploadTokenList)
	tokens.POST("", a.adminUploadTokenCreate)
	tokens.PUT("/:id/disable", a.adminUploadTokenDisable)
}

func (a *App) adminLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	result, err := a.authSvc.Login(body.Username, body.Password, c.ClientIP())
	if err != nil {
		switch err {
		case auth.ErrAccountLocked:
			Fail(c, "ACCOUNT_LOCKED", "account locked")
		case auth.ErrAccountDisabled:
			Fail(c, "UNAUTHORIZED", "account disabled")
		default:
			Fail(c, "UNAUTHORIZED", "invalid credentials")
		}
		return
	}
	OK(c, result)
}

func (a *App) adminRefresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	result, err := a.authSvc.Refresh(body.RefreshToken)
	if err != nil {
		Fail(c, "UNAUTHORIZED", "invalid refresh token")
		return
	}
	OK(c, result)
}

func (a *App) adminLogout(c *gin.Context) {
	token := c.GetString("token")
	a.authSvc.Logout(token)
	OK(c, nil)
}

func (a *App) adminPermissions(c *gin.Context) {
	userId := c.GetString("userId")
	perms, err := a.permCache.Get(userId)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query permissions failed")
		return
	}
	OK(c, gin.H{"permissions": perms})
}

func (a *App) adminSwList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	vs, total, err := a.swRepo.ListAll(page, size)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": vs, "total": total, "page": page, "size": size})
}

func (a *App) adminPluginList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	ps, total, err := a.pluginRepo.ListAll(page, size)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": ps, "total": total, "page": page, "size": size})
}

func (a *App) adminSwPublish(c *gin.Context) {
	file, err := c.FormFile("package")
	if err != nil {
		Fail(c, "INVALID_PARAM", "package file required")
		return
	}
	in := update.PublishInput{
		SwCode:      c.PostForm("swCode"),
		Version:     c.PostForm("version"),
		ReleaseNote: c.PostForm("releaseNote"),
		SwUrl:       c.PostForm("swUrl"),
	}
	if in.SwCode == "" || in.Version == "" {
		Fail(c, "INVALID_PARAM", "swCode and version required")
		return
	}
	src, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "open file failed")
		return
	}
	defer src.Close()
	v, err := a.updateSvc.Publish(c.Request.Context(), in, src, c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "publish failed")
		return
	}
	src2, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "reopen file failed")
		return
	}
	defer src2.Close()
	_ = a.updateSvc.SavePackage(c.Request.Context(), in.SwCode, in.Version, src2)
	OK(c, v)
}

func (a *App) adminSwOffline(c *gin.Context) {
	id := c.Param("id")
	if err := a.updateSvc.Offline(c.Request.Context(), id, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "SW_NOT_FOUND", "not found or invalid status")
		return
	}
	OK(c, nil)
}

func (a *App) adminSwDelete(c *gin.Context) {
	id := c.Param("id")
	confirm := c.Query("confirm") == "true"
	if err := a.updateSvc.Delete(c.Request.Context(), id, confirm, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		if strings.Contains(err.Error(), "confirmation") {
			Fail(c, "CONFIRMATION_REQUIRED", "confirmation required")
		} else {
			Fail(c, "SW_NOT_FOUND", "not found")
		}
		return
	}
	OK(c, nil)
}

func (a *App) adminPluginSubmit(c *gin.Context) {
	file, err := c.FormFile("package")
	if err != nil {
		Fail(c, "INVALID_PARAM", "package file required")
		return
	}
	in := plugin.SubmitInput{
		Name:        c.PostForm("name"),
		Version:     c.PostForm("version"),
		Author:      c.PostForm("author"),
		Category:    c.PostForm("category"),
		Description: c.PostForm("description"),
		DownloadUrl: c.PostForm("downloadUrl"),
	}
	if c.PostForm("keywords") != "" {
		in.Keywords = strings.Split(c.PostForm("keywords"), ",")
	}
	src, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "open file failed")
		return
	}
	defer src.Close()
	p, err := a.pluginSvc.Submit(c.Request.Context(), in, src, c.GetString("userId"), c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "submit failed")
		return
	}
	src2, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "reopen file failed")
		return
	}
	defer src2.Close()
	_ = a.pluginSvc.SavePackage(c.Request.Context(), p.ID, p.Version, src2)
	OK(c, p)
}

func (a *App) adminPluginOffline(c *gin.Context) {
	id := c.Param("id")
	if err := a.pluginSvc.Offline(c.Request.Context(), id, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "PLUGIN_NOT_FOUND", "not found or invalid status")
		return
	}
	OK(c, nil)
}

func (a *App) adminPluginDelete(c *gin.Context) {
	id := c.Param("id")
	confirm := c.Query("confirm") == "true"
	if err := a.pluginSvc.Delete(c.Request.Context(), id, confirm, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		if strings.Contains(err.Error(), "confirmation") {
			Fail(c, "CONFIRMATION_REQUIRED", "confirmation required")
		} else {
			Fail(c, "PLUGIN_NOT_FOUND", "not found")
		}
		return
	}
	OK(c, nil)
}

func (a *App) adminPluginResubmit(c *gin.Context) {
	id := c.Param("id")
	file, err := c.FormFile("package")
	if err != nil {
		Fail(c, "INVALID_PARAM", "package file required")
		return
	}
	in := plugin.SubmitInput{
		Name:        c.PostForm("name"),
		Version:     c.PostForm("version"),
		Author:      c.PostForm("author"),
		Category:    c.PostForm("category"),
		Description: c.PostForm("description"),
		DownloadUrl: c.PostForm("downloadUrl"),
	}
	src, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "open file failed")
		return
	}
	defer src.Close()
	p, err := a.pluginSvc.Resubmit(c.Request.Context(), id, in, src, c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "INVALID_STATUS", "resubmit failed")
		return
	}
	OK(c, p)
}

func (a *App) adminPluginReviewing(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	ps, total, err := a.pluginSvc.ListReviewing(page, size)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": ps, "total": total, "page": page, "size": size})
}

func (a *App) adminPluginReview(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Pass    bool   `json:"pass"`
		Reason  string `json:"reason"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	p, err := a.reviewSvc.Review(review.ReviewInput{
		PluginID: id, Pass: body.Pass, Reason: body.Reason, Comment: body.Comment,
	}, c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "PLUGIN_NOT_FOUND", "review failed")
		return
	}
	OK(c, p)
}

func (a *App) adminIndexRebuild(c *gin.Context) {
	result, err := a.indexSvc.Rebuild(c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "REBUILD_IN_PROGRESS", "rebuild in progress")
		return
	}
	OK(c, result)
}

func (a *App) adminIndexStatus(c *gin.Context) {
	OK(c, a.indexSvc.Status())
}

func (a *App) adminMetrics(c *gin.Context) {
	dimension := c.Query("dimension")
	granularity := c.Query("granularity")
	subjectKey := c.Query("subjectKey")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)
	if !from.IsZero() && !to.IsZero() && to.Sub(from) > 30*24*time.Hour {
		Fail(c, "QUERY_RANGE_TOO_LARGE", "range too large (max 30 days)")
		return
	}
	events, err := a.metricsSvc.Query(dimension, from, to, granularity, subjectKey)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, events)
}

func (a *App) adminAuditList(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)
	if !from.IsZero() && !to.IsZero() && to.Sub(from) > 7*24*time.Hour {
		Fail(c, "QUERY_RANGE_TOO_LARGE", "range too large (max 7 days)")
		return
	}
	action := c.Query("action")
	operator := c.Query("operator")
	targetType := c.Query("targetType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	logs, total, err := a.auditSvc.Query(from, to, action, operator, targetType, page, size)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": logs, "total": total, "page": page, "size": size})
}

func (a *App) adminUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	us, total, err := a.userSvc.ListUsers(page, size)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": us, "total": total, "page": page, "size": size})
}

func (a *App) adminUserCreate(c *gin.Context) {
	var body struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		RoleIDs  []string `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	u, err := a.userSvc.CreateUser(body.Username, body.Password, body.RoleIDs, c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "USER_ALREADY_EXISTS", "user already exists")
		return
	}
	OK(c, u)
}

func (a *App) adminUserGet(c *gin.Context) {
	id := c.Param("id")
	u, err := a.userSvc.GetUser(id)
	if err != nil {
		Fail(c, "USER_NOT_FOUND", "user not found")
		return
	}
	OK(c, u)
}

func (a *App) adminUserDisable(c *gin.Context) {
	id := c.Param("id")
	if err := a.userSvc.Disable(id, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "PROTECTED_USER", "cannot disable")
		return
	}
	OK(c, nil)
}

func (a *App) adminUserEnable(c *gin.Context) {
	id := c.Param("id")
	if err := a.userSvc.Enable(id, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "USER_NOT_FOUND", "user not found")
		return
	}
	OK(c, nil)
}

func (a *App) adminUserDelete(c *gin.Context) {
	id := c.Param("id")
	confirm := c.Query("confirm") == "true"
	if err := a.userSvc.Delete(id, confirm, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		if strings.Contains(err.Error(), "confirmation") {
			Fail(c, "CONFIRMATION_REQUIRED", "confirmation required")
		} else {
			Fail(c, "PROTECTED_USER", "cannot delete")
		}
		return
	}
	OK(c, nil)
}

func (a *App) adminUserAssignRoles(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		RoleIDs []string `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	if err := a.userSvc.AssignRoles(id, body.RoleIDs, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "INTERNAL_ERROR", "assign failed")
		return
	}
	OK(c, nil)
}

func (a *App) adminRoleList(c *gin.Context) {
	roles, err := a.userSvc.ListRoles()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, roles)
}

func (a *App) adminRoleCreate(c *gin.Context) {
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		Description string   `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	r, err := a.userSvc.CreateRole(body.Name, body.Permissions, body.Description, c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "create failed")
		return
	}
	OK(c, r)
}

func (a *App) adminRoleUpdate(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	if err := a.userSvc.UpdateRole(id, body.Permissions, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "INTERNAL_ERROR", "update failed")
		return
	}
	OK(c, nil)
}

func (a *App) adminRoleDelete(c *gin.Context) {
	id := c.Param("id")
	if err := a.userSvc.DeleteRole(id, c.GetString("username"), c.GetString("requestId"), c.ClientIP()); err != nil {
		Fail(c, "INVALID_STATUS", "cannot delete builtin role")
		return
	}
	OK(c, nil)
}

func (a *App) adminConfigList(c *gin.Context) {
	items, err := a.configSvc.List()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, items)
}

func (a *App) adminConfigUpdate(c *gin.Context) {
	key := c.Param("key")
	var body struct {
		Value   string `json:"value"`
		Confirm bool   `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	item, err := a.configSvc.Update(key, body.Value, body.Confirm, c.GetString("username"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		switch err {
		case configsvc.ErrConfigNotFound:
			Fail(c, "CONFIG_NOT_FOUND", "config not found")
		case configsvc.ErrInvalidConfigValue:
			Fail(c, "INVALID_CONFIG_VALUE", "invalid value")
		case configsvc.ErrConfirmationRequired:
			Fail(c, "CONFIRMATION_REQUIRED", "confirmation required")
		default:
			Fail(c, "INTERNAL_ERROR", "update failed")
		}
		return
	}
	OK(c, item)
}

func parseRange(rangeHeader string, totalSize int64) (int64, int64, bool) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	offset, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	var length int64
	if parts[1] == "" {
		length = totalSize - offset
	} else {
		end, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		length = end - offset + 1
	}
	if offset < 0 || offset >= totalSize || length <= 0 {
		return 0, 0, false
	}
	return offset, length, true
}
func (a *App) adminDocsRegenerate(c *gin.Context) {
	apidoc.GetDefaultCache().Rebuild()
	spec := apidoc.GetDefaultCache().GetSpec()
	pathCount := 0
	if spec != nil {
		pathCount = len(spec.Paths)
	}
	OK(c, gin.H{"success": true, "pathCount": pathCount})
}
func (a *App) registerUploadRoutes(g *gin.RouterGroup) {
	uploadGroup := g.Group("/upload")
	uploadGroup.Use(middleware.UploadTokenAuthMiddleware(a.uploadTokenSvc))
	uploadGroup.Use(middleware.UploadRateLimitMiddleware(a.uploadRateLimiter, a.cfg))
	uploadGroup.POST("/plugin", a.uploadPlugin)
	uploadGroup.POST("/software", a.uploadSoftware)
	uploadGroup.POST("/file", a.uploadFile)
}

func (a *App) serveFile(c *gin.Context) {
	path := c.Param("path")
	storedPath := "file/" + strings.TrimPrefix(path, "/")

	stat, err := a.binaryStore.Stat(c.Request.Context(), storedPath)
	if err != nil {
		Fail(c, "FILE_NOT_FOUND", "file not found")
		return
	}

	reader, err := a.binaryStore.Read(c.Request.Context(), storedPath)
	if err != nil {
		Fail(c, "FILE_NOT_FOUND", "file not found")
		return
	}
	defer reader.Close()

	buf := make([]byte, 512)
	n, _ := reader.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	c.DataFromReader(http.StatusOK, stat, contentType, io.MultiReader(bytes.NewReader(buf[:n]), reader), nil)
}

func (a *App) uploadPlugin(c *gin.Context) {
	file, err := c.FormFile("package")
	if err != nil {
		Fail(c, "INVALID_PARAM", "package file required")
		return
	}
	in := uploadsvc.PluginUploadInput{
		Name:        c.PostForm("name"),
		Version:     c.PostForm("version"),
		Author:      c.PostForm("author"),
		Category:    c.PostForm("category"),
		Description: c.PostForm("description"),
		Filename:    file.Filename,
		FileSize:    file.Size,
	}
	if c.PostForm("keywords") != "" {
		in.Keywords = strings.Split(c.PostForm("keywords"), ",")
	}
	src, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "open file failed")
		return
	}
	defer src.Close()

	result, err := a.pluginUploadSvc.Upload(c.Request.Context(), in, src, c.GetString("uploaderIdentity"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		handleUploadError(c, err)
		return
	}
	OK(c, result)
}

func (a *App) uploadSoftware(c *gin.Context) {
	file, err := c.FormFile("package")
	if err != nil {
		Fail(c, "INVALID_PARAM", "package file required")
		return
	}
	in := uploadsvc.SoftwareUploadInput{
		SwCode:      c.PostForm("swCode"),
		Version:     c.PostForm("version"),
		ReleaseNote: c.PostForm("releaseNote"),
		SwUrl:       c.PostForm("swUrl"),
		Filename:    file.Filename,
		FileSize:    file.Size,
	}
	src, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "open file failed")
		return
	}
	defer src.Close()

	result, err := a.softwareUploadSvc.Upload(c.Request.Context(), in, src, c.GetString("uploaderIdentity"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		handleUploadError(c, err)
		return
	}
	OK(c, result)
}

func (a *App) uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, "INVALID_PARAM", "file required")
		return
	}
	in := uploadsvc.FileUploadInput{
		Purpose:  c.PostForm("purpose"),
		Filename: file.Filename,
		FileSize: file.Size,
	}
	src, err := file.Open()
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "open file failed")
		return
	}
	defer src.Close()

	result, err := a.fileUploadSvc.Upload(c.Request.Context(), in, src, c.GetString("uploaderIdentity"), c.GetString("requestId"), c.ClientIP())
	if err != nil {
		handleUploadError(c, err)
		return
	}
	OK(c, result)
}

func handleUploadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, uploadsvc.ErrInvalidParam):
		Fail(c, "INVALID_PARAM", err.Error())
	case errors.Is(err, uploadsvc.ErrInvalidFileType):
		Fail(c, "INVALID_FILE_TYPE", err.Error())
	case errors.Is(err, uploadsvc.ErrFileTooLarge):
		Fail(c, "FILE_TOO_LARGE", err.Error())
	case errors.Is(err, uploadsvc.ErrPluginAlreadyExists):
		Fail(c, "PLUGIN_ALREADY_EXISTS", err.Error())
	case errors.Is(err, uploadsvc.ErrVersionAlreadyExists):
		Fail(c, "VERSION_ALREADY_EXISTS", err.Error())
	case errors.Is(err, uploadsvc.ErrInvalidFilename):
		Fail(c, "INVALID_FILENAME", err.Error())
	default:
		Fail(c, "INTERNAL_ERROR", "upload failed")
	}
}

func (a *App) adminUploadTokenList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	tokens, total, err := a.uploadTokenSvc.List(page, size)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "query failed")
		return
	}
	OK(c, gin.H{"items": tokens, "total": total, "page": page, "size": size})
}

func (a *App) adminUploadTokenCreate(c *gin.Context) {
	var body struct {
		Name             string   `json:"name"`
		UploaderIdentity string   `json:"uploaderIdentity"`
		Permissions      []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, "INVALID_PARAM", "invalid body")
		return
	}
	plaintextToken, record, err := a.uploadTokenSvc.Create(body.Name, body.UploaderIdentity, body.Permissions)
	if err != nil {
		Fail(c, "INTERNAL_ERROR", "create failed")
		return
	}
	_ = a.auditSvc.Record(a.db, audit.Record{
		Operator:   c.GetString("username"),
		Action:     model.AuditActionTokenCreated,
		TargetType: model.AuditTargetUploadToken,
		TargetID:   record.ID,
		Result:     model.AuditResultSuccess,
		Detail:     "create token " + body.Name,
		SourceIP:   c.ClientIP(),
		RequestID:  c.GetString("requestId"),
	})
	OK(c, gin.H{
		"token":       plaintextToken,
		"id":          record.ID,
		"name":        record.Name,
		"permissions": record.Permissions,
		"status":      record.Status,
		"createdAt":   record.CreatedAt,
	})
}

func (a *App) adminUploadTokenDisable(c *gin.Context) {
	id := c.Param("id")
	if err := a.uploadTokenSvc.Disable(id); err != nil {
		Fail(c, "TOKEN_NOT_FOUND", "token not found")
		return
	}
	_ = a.auditSvc.Record(a.db, audit.Record{
		Operator:   c.GetString("username"),
		Action:     model.AuditActionTokenDisabled,
		TargetType: model.AuditTargetUploadToken,
		TargetID:   id,
		Result:     model.AuditResultSuccess,
		Detail:     "disable token " + id,
		SourceIP:   c.ClientIP(),
		RequestID:  c.GetString("requestId"),
	})
	OK(c, nil)
}
