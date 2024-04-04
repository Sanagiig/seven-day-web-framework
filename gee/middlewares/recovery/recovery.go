package recovery

import (
	"common/console"
	"fmt"
	"gee/context"
	"net/http"
	"runtime"
	"strings"
)

func trace(msg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var sb strings.Builder
	sb.WriteString(msg + "\nTraceback")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}

func Recovery() context.HandlerFunc {
	return func(c *context.Context) {
		console.Info("into recover")
		defer func() {
			if err := recover(); err != nil {
				console.Info("recover")
				msg := fmt.Sprintf("%s", err)
				console.Error("%s\n\n", trace(msg))
				c.Fail(http.StatusInternalServerError, "server err")
			}
		}()
	}
}
