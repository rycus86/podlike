package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/loader"
)

func (tc *transformConfiguration) getServiceIndex() int {
	for idx, service := range tc.Session.Project.Services {
		if service.Name == tc.Service.Name {
			return idx
		}
	}

	return -1
}

func (tc *transformConfiguration) getServiceConfig() map[string]interface{} {
	if asMap, err := loader.ParseYAML([]byte(convertToYaml(tc.Service))); err != nil {
		panic(fmt.Sprintf("failed to get service config for %s: %s", tc.Service.Name, err))
	} else {
		return asMap
	}
}
