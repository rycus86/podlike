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
	schemaData, err := ioutil.ReadFile("../testdata/config_schema_v2.4.json")
	if err != nil {
		t.Fatal(err)
	}

	var schema map[string]interface{}

	if err := json.Unmarshal(schemaData, &schema); err != nil {
		t.Fatal(err)
	}

	var (
		fieldRe       = regexp.MustCompile(".* field (.+) not found in type engine.Component")
		healthcheckRe = regexp.MustCompile(".* field (.+) not found in type engine.Healthcheck")
		blkioRe       = regexp.MustCompile(".* field (.+) not found in type engine.BlkioConfig")
	)

	allDefinitions := schema["definitions"].(map[string]interface{})
	serviceDefinition := allDefinitions["service"].(map[string]interface{})
	serviceProperties := serviceDefinition["properties"].(map[string]interface{})

	testProperties := "image: testing\n"

	iterProperties(serviceProperties, allDefinitions, "", &testProperties)

	err = yaml.UnmarshalStrict([]byte(testProperties), &Component{})

	if err != nil {
		yamlErrors := err.(*yaml.TypeError).Errors

		unsupported := make([]string, 0, len(yamlErrors))

		for _, e := range yamlErrors {
			if fieldRe.MatchString(e) {
				unsupported = append(unsupported, fieldRe.ReplaceAllString(e, "$1"))
			} else if healthcheckRe.MatchString(e) {
				unsupported = append(unsupported, healthcheckRe.ReplaceAllString(e, "healthcheck.$1"))
			} else if blkioRe.MatchString(e) {
				unsupported = append(unsupported, blkioRe.ReplaceAllString(e, "blkio_config.$1"))
			} else {
				unsupported = append(unsupported, "? ("+e+")")
			}
		}

		sort.Sort(sort.StringSlice(unsupported))

		fmt.Println("Currently unsupported Compose properties:\n  ", unsupported)

		expectedDescription := "## Unsupported properties\n\n"
		expectedPattern := expectedDescription

		for _, key := range unsupported {
			expectedDescription += "- `" + key + "`\n"
			expectedPattern += "- `" + key + "`.*\n"
		}

		readmeData, err := ioutil.ReadFile("../README.md")
		if err != nil {
			t.Fatal(err)
		}

		readme := string(readmeData)

		if !regexp.MustCompile(expectedPattern).MatchString(readme) {
			t.Error("The list of unsupported properties is not found in the README")
			fmt.Println(expectedDescription)
		}
	} else {
		t.Fatal("The YAML unmarshalling is expected to fail")
	}
}

func iterProperties(properties map[string]interface{}, definitions map[string]interface{}, prefix string, target *string) {
	for key, value := range properties {
		if prefix == "" && key == "image" {
			continue
		}

		*target += prefix + key + ":\n"

		if child, ok := value.(map[string]interface{}); ok {
			processEmbedded(child, definitions, prefix+"  ", target)
		}
	}
}

func processEmbedded(child map[string]interface{}, definitions map[string]interface{}, prefix string, target *string) {
	if reference, ok := child["$ref"]; ok {
		if referenceID, ok := reference.(string); ok {
			processReference(referenceID, definitions, prefix, target)
		}
	}

	if oneOf, ok := child["oneOf"]; ok {
		options := oneOf.([]interface{})

		for _, option := range options {
			if embeddedOption, ok := option.(map[string]interface{}); ok {
				processEmbedded(embeddedOption, definitions, prefix, target)
			}
		}
	}

	if child["type"] == "array" {
		if items, ok := child["items"]; ok {
			if itemsDef, ok := items.(map[string]interface{}); ok {
				*target += prefix + "-\n"

				processEmbedded(itemsDef, definitions, prefix+"  ", target)
			}
		}
	}

	if child["type"] != "object" {
		return
	}

	if props, ok := child["properties"]; ok {
		if properties, ok := props.(map[string]interface{}); ok {
			iterProperties(properties, definitions, prefix, target)
		}
	}

	if props, ok := child["patternProperties"]; ok {
		*target += prefix + "x:\n"

		if properties, ok := props.(map[string]interface{}); ok {
			for _, value := range properties {
				if embedded, ok := value.(map[string]interface{}); ok {
					processEmbedded(embedded, definitions, prefix+"  ", target)
				}
			}
		}
	}
}

func processReference(id string, definitions map[string]interface{}, prefix string, target *string) {
	for key, value := range definitions {
		def, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		parts := strings.Split(id, "/")

		if def["id"] != id && key != parts[len(parts)-1] {
			continue
		}

		if props, ok := def["properties"]; ok {
			if properties, ok := props.(map[string]interface{}); ok {
				iterProperties(properties, definitions, prefix, target)
			}
		}
	}
}
