package main

import (
	"github.com/gonyyi/reqtest"
)

func main() {
	ignore := []string{
		"/favicon.ico",
	}
	err := reqtest.Run(":8080", "test", 20, ignore, nil)
	if err != nil {
		println(err.Error())
	}
}
