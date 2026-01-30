package main

import (
	"os"

	"github.com/jameskingston/k2l/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
