package component

import (
	"errors"
	"fmt"
	"github.com/rycus86/podlike/pkg/convert"
	"io/ioutil"
	"strings"
)

func variablesFromEnvFiles(files interface{}) (map[string]string, error) {
	variables := map[string]string{}

	if files == nil {
		return variables, nil
	}

	if path, ok := files.(string); ok {
		if fromFile, err := fromEnvFile(path); err != nil {
			return nil, err
		} else {
			for key, value := range fromFile {
				variables[key] = value
			}
		}
	}

	if list, ok := files.([]interface{}); ok {
		for _, file := range list {
			if path, ok := file.(string); ok {
				if fromFile, err := fromEnvFile(path); err != nil {
					return nil, err
				} else {
					for key, value := range fromFile {
						variables[key] = value
					}
				}
			} else {
				return nil, errors.New(fmt.Sprintf("wrong env file path: %+v (%T)", file, file))
			}
		}

		return variables, nil
	}

	return nil, errors.New(fmt.Sprintf("unexpected env file(s): %+v (%T)", files, files))
}

func fromEnvFile(path string) (map[string]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	variables := map[string]string{}

	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			variables[parts[0]] = parts[1]
		} else {
			// TODO is this an error with Compose?
		}
	}

	return variables, nil
}

func mergeEnvVariables(fromFiles, fromVariables map[string]string) []string {
	merged := map[interface{}]interface{}{}

	for key, value := range fromFiles {
		merged[key] = value
	}

	for key, value := range fromVariables {
		merged[key] = value
	}

	asSlice, err := convert.ToStringSlice(merged)
	if err != nil {
		panic(nil) // shouldn't happen
	}

	return asSlice
}
