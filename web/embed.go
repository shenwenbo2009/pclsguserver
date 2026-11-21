package web

import (
	"embed"
	"io/fs"
)

//go:embed all:admin
var adminFS embed.FS

//go:embed all:market
var marketFS embed.FS

//go:embed all:swagger-ui
var swaggerUIFS embed.FS

func AdminFS() fs.FS {
	sub, err := fs.Sub(adminFS, "admin")
	if err != nil {
		panic(err)
	}
	return sub
}

func MarketFS() fs.FS {
	sub, err := fs.Sub(marketFS, "market")
	if err != nil {
		panic(err)
	}
	return sub
}
func SwaggerUIFS() fs.FS {
	sub, err := fs.Sub(swaggerUIFS, "swagger-ui")
	if err != nil {
		panic(err)
	}
	return sub
}
