package main

import (
	"fmt"
	"gee"
	"gee/context"
	"gee/middlewares/logger"
	"gee/middlewares/recovery"
	"net/http"
)

func main() {
	r := gee.New()

	r.GET("/index", func(c *context.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.Use(logger.Logger(), recovery.Recovery())
		v1.GET("/", func(c *context.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *context.Context) {
			// expect /hello?name=geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		r.Use(logger.Logger(), recovery.Recovery())
		v2.GET("/hello/:name", func(c *context.Context) {
			str := []byte{1, 2}
			fmt.Println(str[1000])
			//panic(errors.New("manual err"))
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *context.Context) {
			c.Json(http.StatusOK, gee.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}

	v3 := r.Group("/static")
	{
		v3.Static("/", "../static/")
	}
	r.Run(":9999")
}
