package main

import (
	"github.com/rycus86/podlike/pkg/mesh"
	"os"
)

func main() {
	mesh.StartMeshController(os.Args[1:]...)
}
