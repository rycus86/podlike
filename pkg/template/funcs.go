package template

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

func yamlFunc(v interface{}) (string, error) {
	if contents, err := yaml.Marshal(v); err != nil {
		return "", err
	} else {
		return string(contents), nil
	}
}

func indentFunc(indent int, s string) string {
	output := ""
	for _, line := range strings.Split(s, "\n") {
		output += strings.Repeat(" ", indent) + string(line) + "\n"
	}
	return strings.TrimSuffix(output, "\n")
}

func emptyFunc(v interface{}) bool {
	return reflect.ValueOf(v).Len() == 0
}

func notEmptyFunc(v interface{}) bool {
	return !emptyFunc(v)
}

func containsFunc(item, target string) bool {
	return strings.Contains(target, item)
}

func startsWithFunc(item, target string) bool {
	return strings.HasPrefix(target, item)
}

func replaceFunc(oldStr, newStr string, n int, source string) string {
	return strings.Replace(source, oldStr, newStr, n)
}
