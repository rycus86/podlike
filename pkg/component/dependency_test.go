package component

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestDependenciesAsSlice(t *testing.T) {
	deps, err := parseDependencies(`
depends_on:
  - first
  - second`)
	if err != nil {
		t.Fatal(err)
	}

	verifyMatches(t, deps, Dependency{Name: "first"}, Dependency{Name: "second"})
}

func TestDependenciesAsMap(t *testing.T) {
	deps, err := parseDependencies(`
depends_on:
  first:
  second:
    condition: service_started
  third:
    condition: service_healthy`)
	if err != nil {
		t.Fatal(err)
	}

	verifyMatches(t, deps,
		Dependency{Name: "first"},
		Dependency{Name: "second", NeedsHealthyState: false},
		Dependency{Name: "third", NeedsHealthyState: true})
}

func verifyMatches(t *testing.T, actual []Dependency, expected ...Dependency) {
	if len(actual) != len(expected) {
		t.Error(
			"The number of actual dependencies doesn't match the expectations:",
			len(actual), "!=", len(expected))
	}

	for _, exp := range expected {
		found := false

		for _, act := range actual {
			if exp.Name == act.Name {
				found = true

				if exp.NeedsHealthyState != act.NeedsHealthyState {
					t.Error("Healthy state expectation doesn't match for", exp.Name)
				}
			}
		}

		if !found {
			t.Error("Expected dependency not found:", exp.Name)
		}
	}
}

func parseDependencies(yamlConfig string) ([]Dependency, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
		return nil, err
	}

	component := &Component{DependsOn: config["depends_on"]}
	return component.GetDependencies()
}
