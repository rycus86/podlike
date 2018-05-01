package engine

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestSchema(t *testing.T) {
	schema_data, err := ioutil.ReadFile("../testdata/config_schema_v2.4.json")
	if err != nil {
		t.Fatal(err)
	}

	var schema map[string]interface{}

	if err := json.Unmarshal(schema_data, &schema); err != nil {
		t.Fatal(err)
	}

	allDefinitions := schema["definitions"].(map[string]interface{})
	serviceDefinition := allDefinitions["service"].(map[string]interface{})
	serviceProperties := serviceDefinition["properties"].(map[string]interface{})

	testProperties := "image: testing\n"

	for k, _ := range serviceProperties {
		if k == "image" {
			continue
		}

		testProperties = testProperties + k + ":\n"

		// TODO embedded properties, like for healthcheck
		// TODO types? map, list, number, etc.
	}

	err = yaml.UnmarshalStrict([]byte(testProperties), &Component{})

	if err != nil {
		yamlErrors := err.(*yaml.TypeError).Errors
		fieldRe := regexp.MustCompile(".* field (.+) not found .*")

		unsupported := make([]string, 0, len(yamlErrors))

		for _, e := range yamlErrors {
			field := fieldRe.ReplaceAllString(e, "$1")
			unsupported = append(unsupported, field)
		}

		sort.Sort(sort.StringSlice(unsupported))

		fmt.Println("Unsupported Compose properties:", unsupported)

		expectedDescription := "## Unsupported properties\n\n"

		for _, key := range unsupported {
			expectedDescription = expectedDescription + "- `" + key + "`\n"
		}

		readme_data, err := ioutil.ReadFile("../README.md")
		if err != nil {
			t.Fatal(err)
		}

		readme := string(readme_data)

		if !strings.Contains(readme, expectedDescription) {
			t.Error("The list of unsupported properties is not found in the README")
			fmt.Println(expectedDescription)
		}
	} else {
		t.Fatal("The YAML unmarshalling is expected to fail")
	}
}
