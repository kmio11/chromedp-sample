package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed public/*
var content embed.FS

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	api := e.Group("/api")
	api.GET("/hello", helloHandler)

	// Serve embedded static files
	fsys, err := fs.Sub(content, "public")
	if err != nil {
		e.Logger.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(fsys))
	e.GET("/callback", echo.WrapHandler(http.StripPrefix("/callback", fileServer))) // serve index.html at /callback
	e.GET("/*", echo.WrapHandler(fileServer))                                       // serve all other static files

	e.Logger.Fatal(e.Start(":3000"))
}

func helloHandler(c echo.Context) error {
	type HelloResponse struct {
		Message string `json:"message"`
	}

	authHeader := c.Request().Header.Get("Authorization")
	log.Println(authHeader)

	return c.JSON(http.StatusOK, HelloResponse{Message: "hello"})
}
