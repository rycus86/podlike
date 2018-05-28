package engine

import (
	"errors"
	"fmt"
)

func (c *Component) GetDependencies() ([]Dependency, error) {
	if c.DependsOn == nil {
		return []Dependency{}, nil
	}

	if asSlice, ok := c.DependsOn.([]interface{}); ok {
		dependencies := make([]Dependency, len(asSlice), len(asSlice))

		for idx, name := range asSlice {
			if _, ok := name.(string); !ok {
				return nil, errors.New(fmt.Sprintf("string dependency expected: %+v (%T)", name, name))
			}

			dependencies[idx] = Dependency{Name: name.(string), NeedsHealthyState: false}
		}

		return dependencies, nil
	}

	if asMap, ok := c.DependsOn.(map[interface{}]interface{}); ok {
		dependencies := make([]Dependency, 0, len(asMap))

		for name, configuration := range asMap {
			if _, ok := name.(string); !ok {
				return nil, errors.New(fmt.Sprintf("string dependency expected: %+v (%T)", name, name))
			}

			needsHealthyState := false

			if configMap, ok := configuration.(map[interface{}]interface{}); ok {
				if condition, ok := configMap["condition"]; ok {
					if condition == "service_healthy" {
						needsHealthyState = true
					} else if condition != "service_started" {
						return nil, errors.New(fmt.Sprintf(
							"invalid dependency condition: %+v (%T)", condition, condition))
					}
				}
			}

			dependencies = append(dependencies, Dependency{Name: name.(string), NeedsHealthyState: needsHealthyState})
		}

		return dependencies, nil
	}

	return nil, errors.New(fmt.Sprintf("invalid depends_on definition: %+v", c.DependsOn))
}
