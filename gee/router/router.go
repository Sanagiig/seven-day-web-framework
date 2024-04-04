package router

import (
	"gee/context"
	"gee/trie"
	"log"
	"net/http"
	"strings"
)

type HandlerFunc = context.HandlerFunc

type Router struct {
	roots          map[string]*trie.Node
	handlers       map[string]HandlerFunc
	middlewaresMap map[string][]HandlerFunc
}

func New() *Router {
	return &Router{roots: make(map[string]*trie.Node), handlers: make(map[string]HandlerFunc), middlewaresMap: make(map[string][]HandlerFunc)}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0, len(vs))
	for _, part := range vs {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break
			}
		}
	}

	return parts
}

func (r *Router) AddRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern

	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = trie.New()
	}

	r.roots[method].Insert(pattern, parts, 0)

	_, ok = r.handlers[key]
	if ok {
		log.Fatalf("%s is exist", key)
	}
	r.handlers[key] = handler
}

func (r *Router) AddMiddlewares(pattern string, middlewares ...HandlerFunc) {
	m, ok := r.middlewaresMap[pattern]
	if !ok {
		m = []HandlerFunc{}
	}

	r.middlewaresMap[pattern] = append(m, middlewares...)
}

func (r *Router) GetMiddlewares(parts []string) []HandlerFunc {
	middlewares := make([]HandlerFunc, 0)
	key := ""
	for _, part := range parts {
		key += "/" + part
		m, ok := r.middlewaresMap[key]
		if ok {
			middlewares = append(middlewares, m...)
		}
	}
	return middlewares
}

func (r *Router) GetRouterNode(method string, searchParts []string) (node *trie.Node) {
	root, ok := r.roots[method]
	if ok {
		node = root.Search(searchParts, 0)
	}
	return
}

func (r *Router) GetRouter(method string, path string) (*trie.Node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)

	n := r.GetRouterNode(method, searchParts)
	if n != nil {
		parts := parsePattern(n.Pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}

			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *Router) Handle(c *context.Context) {
	n, params := r.GetRouter(c.Method, c.Path)

	if n != nil {
		key := c.Method + "-" + n.Pattern
		middlewares := r.GetMiddlewares(parsePattern(n.Pattern))
		middlewares = append(middlewares, r.handlers[key])
		c.Params = params
		c.Handlers = middlewares
		c.Next()
	} else {
		c.String(http.StatusNotFound, "404 Not Found %s\n", c.Path)
	}
}
