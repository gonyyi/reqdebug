package main

import (
	"github.com/gonyyi/reqdebug"
)

func main() {
	ignore := []string{
		"/favicon.ico",
	}
	err := reqdebug.Run(":10174", "test", 5, ignore, nil)
	if err != nil {
		println(err.Error())
	}
}
