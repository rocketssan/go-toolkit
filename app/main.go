package main

import (
	"fmt"

	"github.com/rocketssan/toolkit"
)

func main() {
	var tools toolkit.Tools

	s := tools.RandomString(10)
	fmt.Println("Randome string:", s)
}
