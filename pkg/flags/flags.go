package flags

import (
	"flag"
	"fmt"
	"github.com/rycus86/podlike/pkg/config"
	"github.com/rycus86/podlike/pkg/template"
	"github.com/rycus86/podlike/pkg/version"
	"os"
)

var (
	pids, volumes, logs, pull bool
)

func init() {
	setupVariables()
}

func setupVariables() {
	flag.BoolVar(&pids, "pids", true, "Enable (default) or disable PID sharing")
	flag.BoolVar(&volumes, "volumes", false, "Enable volume sharing from the controller")
	flag.BoolVar(&logs, "logs", false, "Stream logs from the components")
	flag.BoolVar(&pull, "pull", false, "Always pull the images for the components when starting")
}

func Parse() *config.Configuration {
	if len(os.Args) > 1 {
		if os.Args[1] == "template" {
			processTemplate()
		} else if os.Args[1] == "version" {
			printVersion()
		}
	}

	flag.Parse()

	if flag.NArg() > 0 {
		panic(fmt.Sprintf("Invalid command line argument: %s", flag.Arg(0)))
	}

	return &config.Configuration{
		SharePids:    pids,
		ShareVolumes: volumes,
		StreamLogs:   logs,
		AlwaysPull:   pull,
	}
}

func processTemplate() {
	template.PrintTemplatedStack(os.Args[2:]...)
	os.Exit(0)
}

func printVersion() {
	v := version.Parse()

	fmt.Println(v.StringForCommandLine())
	os.Exit(0)
}
