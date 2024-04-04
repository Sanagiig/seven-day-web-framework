package logger

import (
	"common/console"
	"gee"
	"gee/context"
	"time"
)

func Logger() gee.HandlerFunc {
	return func(c *context.Context) {
		t := time.Now()
		c.Next()
		console.Info("<%s> [%d] %s in %v", c.Method, c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
