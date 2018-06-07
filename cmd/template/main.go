package main

import (
	"github.com/rycus86/podlike/pkg/template"
	"os"
)

func main() {
	template.PrintTemplatedStack(os.Args[1:]...)
}
