package template

import "fmt"

func (tc *transformConfiguration) getServiceIndex() int {
	for idx, service := range tc.Session.Project.Services {
		if service.Name == tc.Service.Name {
			return idx
		}
	}

	return -1
}

func (tc *transformConfiguration) getServiceConfig() map[string]interface{} {
	for _, configFile := range tc.Session.ConfigFiles {
		if rawServices, ok := configFile.Config["services"]; !ok {
			continue
		} else {
			services := rawServices.(map[string]interface{})
			if svc, ok := services[tc.Service.Name]; ok {
				return svc.(map[string]interface{})
			}
		}
	}

	panic(fmt.Sprintf("service not found: %s\n%+v", tc.Service.Name, tc.Session.ConfigFiles))
}
