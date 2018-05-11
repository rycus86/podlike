package engine

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"sort"
	"testing"
)

func ExampleCopyConfigStyles() {
	fmt.Printf("%v\n", parseAndVerify(nil, `
pod.copy.test: /simple/source:/to/target
`))
	fmt.Printf("%v\n", parseAndVerify(nil, `
pod.copy.test: >
  /etc/conf/logging.conf:/opt/legacy/app/conf/logs.xml
`))
	fmt.Printf("%v\n", parseAndVerify(nil, `
pod.copy.test: |
  - /etc/conf/ssl.cert:/etc/ssl/myapp.cert
  - /etc/conf/ssl.key:/etc/ssl/myapp.key
`))
	fmt.Printf("%v\n", sorted(parseAndVerify(nil, `
pod.copy.test: |
  /etc/conf/cache.conf: /opt/cache/conf.d/default.conf
  /etc/conf/ssl.cert:   /etc/ssl/myapp.cert
  /etc/conf/ssl.key:    /etc/ssl/myapp.key
`)))

	// Output:
	// [{/simple/source /to/target}]
	// [{/etc/conf/logging.conf /opt/legacy/app/conf/logs.xml}]
	// [{/etc/conf/ssl.cert /etc/ssl/myapp.cert} {/etc/conf/ssl.key /etc/ssl/myapp.key}]
	// [{/etc/conf/cache.conf /opt/cache/conf.d/default.conf} {/etc/conf/ssl.cert /etc/ssl/myapp.cert} {/etc/conf/ssl.key /etc/ssl/myapp.key}]
}

func ExampleParsingFailures() {
	parseAndExpectFailure("pod.copy.test: /src")
	parseAndExpectFailure("pod.copy.test: /src: /dst")
	parseAndExpectFailure(`
pod.copy.test: >
  /src: /dst: /extra`)
	parseAndExpectFailure(`
pod.copy.test: >
  /not: /multiline
  /so/it: /fails`)
	parseAndExpectFailure(`
pod.copy.test: >
  - /not:/multiline
  - /so/parsed:/wrong`)
	parseAndExpectFailure(`
pod.copy.test: |
  /number: 1`)
	parseAndExpectFailure(`
pod.copy.test: |
  - /valid:/ok
  - /invalid`)
	parseAndExpectFailure(`
pod.copy.test: |
  '': /empty`)
	parseAndExpectFailure(`
pod.copy.test: >
  /should:/have
  /been:/with/pipe
  /not:/greater/than/sign`)

	// Output:
	// invalid pod.copy configuration: /src
	// yaml: mapping values are not allowed in this context
	// yaml: mapping values are not allowed in this context
	// yaml: mapping values are not allowed in this context
	// invalid pod.copy configuration: [/not:/multiline - /so/parsed:/wrong] [/not:/multiline - /so/parsed:/wrong]
	// not a string value: 1 (int)
	// invalid pod.copy configuration: [/valid:/ok /invalid] [/invalid]
	// invalid pod.copy configuration: map[:/empty] [:/empty]
	// invalid pod.copy configuration: /should:/have /been:/with/pipe /not:/greater/than/sign
}

func TestParseAsSimpleString(t *testing.T) {
	parseAndVerify(t, "pod.copy.test: /src:/dst",
		CopyConfig{"/src", "/dst"})

	parseAndVerify(t, `
pod.copy.test: >
  /new:/line
`, CopyConfig{"/new", "/line"})
}

func TestParseAsSequenceOfStrings(t *testing.T) {
	parseAndVerify(t, `
pod.copy.test: | 
  - /one:/first
  - /two:/second
`, CopyConfig{"/one", "/first"}, CopyConfig{"/two", "/second"})

	parseAndVerify(t, `
pod.copy.test: >
  [/one:/first, /two:/second]
`, CopyConfig{"/one", "/first"}, CopyConfig{"/two", "/second"})
}

func TestParseAsMappingOfStrings(t *testing.T) {
	parseAndVerify(t, `
pod.copy.test: | 
  /one: /first
  /two: /second
`, CopyConfig{"/one", "/first"}, CopyConfig{"/two", "/second"})

	parseAndVerify(t, `
pod.copy.test: | 
  /with:        /in
  /some:        /the
  /whitespace:  /mapping
`,
		CopyConfig{"/with", "/in"},
		CopyConfig{"/some", "/the"},
		CopyConfig{"/whitespace", "/mapping"})
}

func TestSameSourceDifferentTargets(t *testing.T) {
	parseAndVerify(t, `
pod.copy.test: |
  - '/src:/target/1'
  - '/src:/target/2'
`, CopyConfig{"/src", "/target/1"}, CopyConfig{"/src", "/target/2"})
}

func parseAndVerify(t *testing.T, yamlConfig string, expectedConfigs ...CopyConfig) []CopyConfig {
	definition, err := getCopyDefinition(yamlConfig)
	if err != nil {
		if t == nil {
			fmt.Println("Invalid YAML:", yamlConfig, err)
			return nil
		}

		t.Fatal("Invalid YAML:", yamlConfig, err)
	}

	parsed, err := parseCopyConfig(definition)
	if err != nil {
		if t == nil {
			fmt.Println("Failed to parse:", definition, err)
			return nil
		}

		t.Fatal("Failed to parse:", definition, err)
	}

	if expectedConfigs == nil {
		return parsed
	}

	if len(parsed) != len(expectedConfigs) {
		if t == nil {
			fmt.Println("Parsed length doesn't match expected:", len(parsed), "!=", len(expectedConfigs), definition)
			return nil
		}

		t.Error("Parsed length doesn't match expected:", len(parsed), "!=", len(expectedConfigs), definition)
	}

	for _, actual := range parsed {
		found := false

		for _, expected := range expectedConfigs {
			if actual.Source == expected.Source && actual.Target == expected.Target {
				found = true
				break
			}
		}

		if !found {
			if t == nil {
				fmt.Println("Parsed item", actual, "not found in", expectedConfigs)
				return nil
			}

			t.Error("Parsed item", actual, "not found in", expectedConfigs)
		}
	}

	return parsed
}

func sorted(configs []CopyConfig) []CopyConfig {
	sort.Slice(configs, func(i, j int) bool { return configs[i].Source < configs[j].Source })
	return configs
}

func parseAndExpectFailure(yamlConfig string) {
	definition, err := getCopyDefinition(yamlConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	parsed, err := parseCopyConfig(definition)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(fmt.Sprintf("parsed as: %+v", parsed))
	}
}

func getCopyDefinition(yamlConfig string) (string, error) {
	var config map[string]string
	if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
		return "", err
	}

	definition, ok := config["pod.copy.test"]
	if !ok {
		return "", errors.New("missing pod.copy.test key in the YAML")
	}

	return definition, nil
}
