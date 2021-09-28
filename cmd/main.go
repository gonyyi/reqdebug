package main

import (
	"github.com/gonyyi/reqtest"
)

func main() {
	ignore := []string{
		"/favicon.ico",
	}

	err := reqtest.Run(":10173", "test123","r", 20, ignore, nil)
	if err != nil {
		println(err.Error())
	}
}
