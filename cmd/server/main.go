package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"pclsguserver/internal/apidoc"
	"pclsguserver/internal/config"
	"pclsguserver/internal/httpapi"
	inithandler "pclsguserver/internal/init"
	"pclsguserver/internal/pkg/logger"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/store"
	"pclsguserver/web"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var needRestart bool

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	logger.Init("info")
	log := logger.L()

	for {
		if inithandler.IsFirstRun(*configPath) {
			log.Info().Msg("first run detected, entering init mode")
			needRestart = false
			runInitMode(*configPath)
			if needRestart {
				log.Info().Msg("initialization completed, restarting...")
				viper.Reset()
				continue
			}
			log.Info().Msg("init mode exited without completion")
			return
		}

		err := runNormalMode(*configPath)
		if err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
		break
	}
}

func runNormalMode(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	log := logger.L()
	log.Info().Str("dataDir", cfg.Server.DataDir).Msg("starting pclsguserver")

	if err := os.MkdirAll(cfg.Server.DataDir, 0o755); err != nil {
		return err
	}

	db, err := repo.InitDB(cfg)
	if err != nil {
		return err
	}
	log.Info().Msg("database initialized")

	bs, err := store.NewFsStore(cfg.Storage.Path)
	if err != nil {
		return err
	}

	srv := httpapi.NewServer(cfg, &httpapi.Deps{
		DB:          db,
		BinaryStore: bs,
	})

	app := httpapi.NewApp(cfg, db, bs)
	app.RegisterRoutes(srv.Engine())

	apidoc.GetDefaultCache().SetVersion("1.0.0")
	apidoc.RegisterSwaggerRoutes(srv.Engine())
	apidoc.RegisterDocsRoutes(srv.Engine())
	registerSwaggerStaticRoutes(srv.Engine())

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.L().Error().Interface("panic", r).Msg("apidoc generation panic")
			}
		}()
		spec := apidoc.GetDefaultCache().GetSpec()
		if spec != nil {
			logger.L().Info().Int("paths", len(spec.Paths)).Msg("api doc generated")
		}
	}()

	registerStaticRoutes(srv.Engine())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("shutdown signal received")
		cancel()
	}()

	log.Info().Int("port", cfg.Server.Port).Msg("server listening")
	return srv.Run(ctx)
}

func runInitMode(configPath string) {
	cfg, err := config.Load("")
	if err != nil {
		logger.L().Fatal().Err(err).Msg("load default config failed")
	}

	srv := httpapi.NewServer(cfg, &httpapi.Deps{})

	initHandler := inithandler.NewInitHandler(configPath)
	initHandler.RegisterInitRoutes(srv.Engine())

	registerInitStaticRoutes(srv.Engine())

	srv.Engine().NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/init")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	logger.L().Info().Int("port", cfg.Server.Port).Msg("init server listening")
	if err := srv.Run(ctx); err != nil && err != http.ErrServerClosed {
		logger.L().Error().Err(err).Msg("init server error")
	}
}

func registerSwaggerStaticRoutes(r *gin.Engine) {
	r.GET("/swagger/swagger-ui.css", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400")
		c.FileFromFS("swagger-ui.css", http.FS(web.SwaggerUIFS()))
	})
	r.GET("/swagger/swagger-ui-bundle.js", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400")
		c.FileFromFS("swagger-ui-bundle.js", http.FS(web.SwaggerUIFS()))
	})
}

func registerInitStaticRoutes(r *gin.Engine) {
	adminStatic := http.FileServer(http.FS(web.AdminFS()))

	r.GET("/init", func(c *gin.Context) {
		c.Request.URL.Path = "/"
		c.Header("Cache-Control", "no-cache")
		adminStatic.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/init/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		if !hasExt(path) {
			c.Request.URL.Path = "/index.html"
		} else {
			c.Request.URL.Path = path
		}
		c.Header("Cache-Control", "no-cache")
		adminStatic.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/admin/assets/*filepath", func(c *gin.Context) {
		c.Request.URL.Path = "/assets" + c.Param("filepath")
		c.Header("Cache-Control", "no-cache")
		adminStatic.ServeHTTP(c.Writer, c.Request)
	})
}

func registerStaticRoutes(r *gin.Engine) {
	adminStatic := http.FileServer(http.FS(web.AdminFS()))
	marketStatic := http.FileServer(http.FS(web.MarketFS()))

	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		marketStatic.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/market", func(c *gin.Context) {
		c.Request.URL.Path = "/"
		c.Header("Cache-Control", "no-cache")
		marketStatic.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/market/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		if !hasExt(path) {
			c.Request.URL.Path = "/index.html"
		} else {
			c.Request.URL.Path = path
		}
		c.Header("Cache-Control", "no-cache")
		marketStatic.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/admin", func(c *gin.Context) {
		c.Request.URL.Path = "/"
		c.Header("Cache-Control", "no-cache")
		adminStatic.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/admin/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		if !hasExt(path) {
			c.Request.URL.Path = "/index.html"
		} else {
			c.Request.URL.Path = path
		}
		c.Header("Cache-Control", "no-cache")
		adminStatic.ServeHTTP(c.Writer, c.Request)
	})
}

func hasExt(path string) bool {
	if path == "" {
		return false
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return i < len(path)-1
		}
		if path[i] == '/' {
			return false
		}
	}
	return false
}
