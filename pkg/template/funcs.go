package template

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

func yamlFunc(v interface{}) (string, error) {
	if contents, err := yaml.Marshal(v); err != nil {
		panic(err)
	} else {
		return string(contents), nil
	}
}

func indentFunc(indent int, s string) (string, error) {
	output := ""
	for _, line := range strings.Split(s, "\n") {
		output += strings.Repeat(" ", indent) + string(line) + "\n"
	}
	return strings.TrimSuffix(output, "\n"), nil
}

func emptyFunc(v interface{}) bool {
	return reflect.ValueOf(v).Len() == 0
}

func notEmptyFunc(v interface{}) bool {
	return !emptyFunc(v)
}
