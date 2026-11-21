package apidoc

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>pclsguserver API</title>
<link rel="stylesheet" href="/swagger/swagger-ui.css">
<style>body{margin:0}</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="/swagger/swagger-ui-bundle.js"></script>
<script>
window.onload=function(){
  SwaggerUIBundle({
    url:'/swagger/doc.json',
    dom_id:'#swagger-ui',
    deepLinking:true,
    presets:[SwaggerUIBundle.presets.apis],
    layout:'BaseLayout'
  });
};
</script>
</body>
</html>`

func RegisterSwaggerRoutes(r *gin.Engine) {
	r.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})

	r.GET("/swagger/doc.json", func(c *gin.Context) {
		spec := GetDefaultCache().GetSpec()
		if spec == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "spec not available"})
			return
		}
		c.JSON(http.StatusOK, spec)
	})
}

func RegisterDocsRoutes(r *gin.Engine) {
	r.GET("/docs/api.md", func(c *gin.Context) {
		md := GetDefaultCache().GetMarkdown()
		if md == "" {
			c.String(http.StatusServiceUnavailable, "文档尚未生成")
			return
		}
		c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(md))
	})
}
