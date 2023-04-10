package main

import (
	"fmt"

	"github.com/fuddle-io/fuddle/pkg/cli"
)

func main() {
	if err := cli.Start(); err != nil {
		fmt.Println(err)
	}
}
