package main

import (
	"github.com/gonyyi/reqdebug"
)

func main() {
	err := reqdebug.Run(":10173", "http test", []string{"/favicon.ico"})
	if err != nil {
		println(err.Error())
	}
}
