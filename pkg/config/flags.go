package config

import (
	"flag"
	"fmt"
	"github.com/rycus86/podlike/pkg/version"
	"os"
)

var (
	pids, volumes, logs, pull bool
)

type Configuration struct {
	SharePids    bool
	ShareVolumes bool
	StreamLogs   bool
	AlwaysPull   bool
}

func init() {
	setupVariables()
}

func setupVariables() {
	flag.BoolVar(&pids, "pids", true, "Enable (default) or disable PID sharing")
	flag.BoolVar(&volumes, "volumes", false, "Enable volume sharing from the controller")
	flag.BoolVar(&logs, "logs", false, "Stream logs from the components")
	flag.BoolVar(&pull, "pull", false, "Always pull the images for the components when starting")
}

func Parse() *Configuration {
	flag.Parse()

	if flag.NArg() > 0 {
		if flag.Arg(0) == "version" {
			v := version.Parse()

			fmt.Println(v.StringForCommandLine())
			os.Exit(0)
		} else {
			panic(fmt.Sprintf("Invalid command line argument: %s", flag.Arg(0)))
		}
	}

	return &Configuration{
		SharePids:    pids,
		ShareVolumes: volumes,
		StreamLogs:   logs,
		AlwaysPull:   pull,
	}
}
