package main

import (
	"github.com/gonyyi/reqdebug"
)

func main() {
	err := reqdebug.Run(":10173", "gon/y/yi")
	if err != nil {
		println(err.Error())
	}
}
