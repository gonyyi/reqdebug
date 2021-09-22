package main

import (
	"github.com/gonyyi/slackbi/cmd/reqdebug"
)

func main() {
	err := reqdebug.Run(":10173", "gon/y/yi")
	if err != nil {
		println(err.Error())
	}
}
