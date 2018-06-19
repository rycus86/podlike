package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/loader"
)

// Remove and return the Compose v2 compatible depends_on configurations.
// These won't parse for the v3 service type, so we'll just have to add them
// back after the conversion is done.
func extractDependsOnConfig(configuration map[string]interface{}) map[string]interface{} {
	servicesWithDependsOn := map[string]interface{}{}

	for name, config := range configuration {
		if mConfig, ok := config.(map[string]interface{}); !ok {
			panic(fmt.Sprintf("unexpected service definition type: %T\n%+v", config, config))

		} else if dependsOn, ok := mConfig["depends_on"]; ok {
			validateDependsOn(dependsOn)

			servicesWithDependsOn[name] = map[string]interface{}{
				"depends_on": dependsOn,
			}

			delete(mConfig, "depends_on")
		}
	}

	return servicesWithDependsOn
}

func validateDependsOn(v interface{}) {
	switch v.(type) {
	case []string:
		// ok

	case []interface{}:
		for _, item := range v.([]interface{}) {
			if _, ok := item.(string); !ok {
				panic(fmt.Sprintf("invalid depends_on list item: %+v %T", item, item))
			}
		}

	case map[string]interface{}:
		for svc, config := range v.(map[string]interface{}) {
			if mConfig, ok := config.(map[string]interface{}); !ok {
				panic(fmt.Sprintf("invalid depends_on defined for %s (type %T)", svc, config))
			} else if condition, ok := mConfig["condition"]; !ok {
				panic(fmt.Sprintf("condition not found for %s dependency : %+v", svc, mConfig))
			} else if condition != "service_started" && condition != "service_healthy" {
				panic(fmt.Sprintf("invalid condition defined for %s : %s", svc, condition))
			}
		}

	default:
		panic(fmt.Sprintf("unexpected depends_on config type: %T\n%+v", v, v))
	}
}

// Merge in the previously removed depends_on configurations to the rendered YAML string.
func insertDependsOnConfig(target string, source map[string]interface{}, service string) string {
	var config interface{}

	if svcConfig, ok := source[service]; !ok {
		return target
	} else if mConfig, ok := svcConfig.(map[string]interface{}); !ok {
		panic(fmt.Sprintf("somehow lost the depends_on settings for %s in %+v", service, svcConfig))
	} else {
		config = mConfig["depends_on"]
	}

	if config == nil {
		panic(fmt.Sprintf("somehow lost the depends_on settings for %s in %+v", service, source))
	}

	parsed, err := loader.ParseYAML([]byte(target))
	if err != nil {
		panic(fmt.Sprintf("failed to parse the output YAML for %s : %s\n%s", service, err.Error(), target))
	}

	parsed["depends_on"] = config

	return convertToYaml(parsed)
}
