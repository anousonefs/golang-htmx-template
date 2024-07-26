package main

import (
	"github.com/anousonefs/golang-htmx-template/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
