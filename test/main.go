package main

import (
	"fmt"
	"flag"
	"github.com/kataras/iris"
	"net/http"
	"os"
)

func main() {
	name := flag.String("name", "", "Input your username")
	department := flag.String("department", "", "Input your department")
	port := flag.String("port", "", "Listen port")
	flag.Parse()

	fmt.Printf("Hi %s from %s!\n", *name, *department)

	iris.Get("/", func(c *iris.Context) {
		c.JSON(http.StatusOK, fmt.Sprintf("Hi %s from %s", *name, *department))
	})
	iris.Listen(fmt.Sprintf(":%s", *port))
}
