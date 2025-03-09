package main

import (
	"log"

	"github.com/rocketssan/toolkit"
)

func main() {
	toSlug := "Watashi ha utyu-jin da. 1 2 3"

	var tools toolkit.Tools

	slugified, err := tools.Slugify(toSlug)
	if err != nil {
		log.Println(err)
	}

	log.Println(slugified)
}
