package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"io/ioutil"
	"os"
	"path"
)

func NewSession(inputFiles ...string) *transformSession {
	if len(inputFiles) == 0 {
		panic("No input files given")
	}

	// special case for reading the input YAML from the standard input
	if len(inputFiles) == 1 && inputFiles[0] == "-" {
		inputFiles = []string{os.Stdin.Name()}
	}

	configFiles := make([]types.ConfigFile, len(inputFiles), len(inputFiles))

	session := &transformSession{
		WorkingDir:     path.Dir(inputFiles[0]),
		Args:           map[string]interface{}{},
		Configurations: map[string]transformConfiguration{},
	}

	for idx, inputFile := range inputFiles {
		contents, err := ioutil.ReadFile(inputFile)
		if err != nil {
			panic(fmt.Sprintf("failed to read the contents of %s : %s", inputFile, err.Error()))
		}

		parsed, err := loader.ParseYAML(contents)
		if err != nil {
			panic(fmt.Sprintf("failed to parse a YAML : %s\n%s", err.Error(), string(contents)))
		}

		configFiles[idx] = types.ConfigFile{
			Filename: path.Base(inputFile),
			Config:   parsed,
		}
	}

	session.prepareConfiguration(configFiles)

	config, err := loader.Load(types.ConfigDetails{
		ConfigFiles: configFiles,
		WorkingDir:  session.WorkingDir,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to load stack YAMLs into a config : %s\nfrom: %+v", err.Error(), configFiles))
	}

	session.Project = config

	for _, svc := range config.Services {
		current := svc

		if cfg, ok := session.Configurations[svc.Name]; ok {
			cfg.Service = &current
			session.Configurations[svc.Name] = cfg
		}
	}

	return session
}

func (ts *transformSession) ReplaceService(svc *types.ServiceConfig) {
	ts.Project.Services = types.Services{*svc}
	cfg := ts.Configurations[svc.Name]
	cfg.Service = svc
	ts.Configurations[svc.Name] = cfg
}
