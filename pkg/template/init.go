package template

import (
	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"io/ioutil"
	"os"
	"path"
)

func newSession(inputFiles ...string) *transformSession {
	if len(inputFiles) == 0 {
		panic("No input files given")
	}

	// special case for reading the input YAML from the standard input
	if len(inputFiles) == 1 && inputFiles[0] == "-" {
		inputFiles = []string{os.Stdin.Name()}
	}

	session := &transformSession{
		WorkingDir:  path.Dir(inputFiles[0]),
		ConfigFiles: make([]types.ConfigFile, len(inputFiles), len(inputFiles)),

		Args: map[string]interface{}{},

		Configurations: map[string]transformConfiguration{},
	}

	for idx, inputFile := range inputFiles {
		contents, err := ioutil.ReadFile(inputFile)
		if err != nil {
			panic(err)
		}

		parsed, err := loader.ParseYAML(contents)
		if err != nil {
			panic(err)
		}

		session.ConfigFiles[idx] = types.ConfigFile{
			Filename: path.Base(inputFile),
			Config:   parsed,
		}
	}

	session.prepareConfiguration()

	config, err := loader.Load(types.ConfigDetails{
		ConfigFiles: session.ConfigFiles,
		WorkingDir:  session.WorkingDir,
	})
	if err != nil {
		panic(err)
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
