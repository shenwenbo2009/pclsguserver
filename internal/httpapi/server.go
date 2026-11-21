package httpapi

import (
	"context"
	"net/http"
	"time"

	"pclsguserver/internal/config"
	"pclsguserver/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    *config.Config
	engine *gin.Engine
	deps   *Deps
}

type Deps struct {
	DB           interface{}
	BinaryStore  interface{}
	IndexBuilder interface{}
}

func NewServer(cfg *config.Config, deps *Deps) *Server {
	if cfg.Server.DataDir == "" {
		cfg.Server.DataDir = "./data"
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	s := &Server{cfg: cfg, engine: engine, deps: deps}
	s.setupMiddleware()
	return s
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) setupMiddleware() {
	s.engine.Use(RecoveryMiddleware())
	s.engine.Use(RequestIDMiddleware())
	s.engine.Use(AccessLogMiddleware())
	s.engine.Use(CORSMiddleware())
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         addrOf(s.cfg.Server.Port),
		Handler:      s.engine,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}
	if srv.ReadTimeout == 0 {
		srv.ReadTimeout = 30 * time.Second
	}
	if srv.WriteTimeout == 0 {
		srv.WriteTimeout = 30 * time.Second
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	return srv.ListenAndServe()
}

func addrOf(port int) string {
	return ":" + itoa(port)
}

func itoa(i int) string {
	if i == 0 {
		return "8080"
	}
	buf := make([]byte, 0, 6)
	if i == 0 {
		return "0"
	}
	digits := [10]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for i > 0 {
		buf = append([]byte{digits[i%10]}, buf...)
		i /= 10
	}
	return string(buf)
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		logger.FromCtx(c.Request.Context()).Error().
			Interface("panic", recovered).
			Str("path", c.Request.URL.Path).
			Msg("panic recovered")
		FailWithStatus(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		c.Abort()
	})
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = NewRequestID()
		}
		c.Set("requestId", rid)
		c.Header("X-Request-Id", rid)
		ctx := logger.WithRequestId(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.FromCtx(c.Request.Context()).Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Int64("latencyMs", latency.Milliseconds()).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-Id, Range, X-Upload-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Range, X-Checksum, ETag")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
