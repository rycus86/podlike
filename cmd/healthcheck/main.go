package main

import (
	"github.com/rycus86/podlike/pkg/healthcheck"
	"os"
)

func main() {
	if healthcheck.Check() {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
