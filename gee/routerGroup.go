package gee

import (
	"gee/context"
	"log"
	"net/http"
	"path"
)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	engine := rg.engine
	newGroup := &RouterGroup{
		prefix: rg.prefix + prefix,
		parent: rg,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (rg *RouterGroup) Use(middlewares ...HandlerFunc) {
	rg.middlewares = append(rg.middlewares, middlewares...)
	rg.engine.router.AddMiddlewares(rg.prefix, middlewares...)
}

func (rg *RouterGroup) AddRoute(method string, comp string, handler HandlerFunc) {
	pattern := rg.prefix + comp
	log.Printf("Add Route %4s - %s", method, pattern)
	rg.engine.router.AddRoute(method, pattern, handler)
}

func (rg *RouterGroup) GET(pattern string, handler HandlerFunc) {
	rg.AddRoute("GET", pattern, handler)
}

func (rg *RouterGroup) POST(pattern string, handler HandlerFunc) {
	rg.AddRoute("POST", pattern, handler)
}

func (rg *RouterGroup) createStaticHandler(relativaPath string, hfs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(rg.prefix, relativaPath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(hfs))

	return func(c *context.Context) {
		file := c.Param("filepath")
		if _, err := hfs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (rg *RouterGroup) Static(relativePath string, root string) {
	handler := rg.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")

	rg.GET(urlPattern, handler)
}
