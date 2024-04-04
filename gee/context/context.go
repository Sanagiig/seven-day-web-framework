package context

import (
	"common/console"
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}
type HandlerFunc func(c *Context)

type Context struct {
	//origin obj
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
	// middleware
	Handlers []HandlerFunc
	index    int
}

func New(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:   w,
		Req:      req,
		Path:     req.URL.Path,
		Method:   req.Method,
		index:    -1,
		Handlers: make([]HandlerFunc, 0),
	}
}

func (c *Context) Next() {
	l := len(c.Handlers)
	if l > 0 && l > c.index {
		c.index++
		c.Handlers[c.index](c)
	} else {
		console.Error("index [%d] out of range [%d]", c.index, len(c.Handlers))
	}
}

func (c *Context) Param(key string) string {
	val, _ := c.Params[key]
	return val
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header()
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) Json(code int, obj interface{}) {
	c.Status(code)
	c.SetHeader("Content-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)

	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.Status(code)
	c.SetHeader("Content-Type", "text/html")
	c.Writer.Write([]byte(html))
}

func (c *Context) Fail(code int, msg string) {
	c.Status(code)
	c.Writer.Write([]byte(msg))
}
