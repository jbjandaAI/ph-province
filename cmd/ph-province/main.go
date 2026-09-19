package main

import (
	"os"

	"ph-province/internal/app"
)

var version = "dev"

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, version))
}
