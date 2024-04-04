package gee

import (
	"gee/context"
	"gee/router"
	"net/http"
	"strings"
)

type Engine struct {
	*RouterGroup
	router *router.Router
	groups []*RouterGroup
}

type HandlerFunc = context.HandlerFunc
type H = map[string]string

func New() *Engine {
	engine := &Engine{router: router.New()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.AddRoute(method, pattern, handler)
}

func (engine *Engine) HEAD(pattern string, handler HandlerFunc) {
	engine.addRoute("HEAD", pattern, handler)
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) PUT(pattern string, handler HandlerFunc) {
	engine.addRoute("PUT", pattern, handler)
}

func (engine *Engine) DELETE(pattern string, handler HandlerFunc) {
	engine.addRoute("DELETE", pattern, handler)
}

func (engine *Engine) CONNECT(pattern string, handler HandlerFunc) {
	engine.addRoute("CONNECT", pattern, handler)
}

func (engine *Engine) OPTIONS(pattern string, handler HandlerFunc) {
	engine.addRoute("OPTIONS", pattern, handler)
}

func (engine *Engine) TRACE(pattern string, handler HandlerFunc) {
	engine.addRoute("TRACE", pattern, handler)
}

func (engine *Engine) PATCH(pattern string, handler HandlerFunc) {
	engine.addRoute("PATCH", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) GetMiddlewares(pattern string) []HandlerFunc {
	middlewares := make([]HandlerFunc, len(engine.groups))
	for _, rg := range engine.groups {
		if strings.HasPrefix(pattern, rg.prefix) {
			middlewares = append(middlewares, rg.middlewares...)
		}
	}
	return middlewares
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.router.Handle(context.New(w, req))
}
