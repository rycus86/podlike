package component

import (
	"github.com/rycus86/podlike/pkg/volume"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestVolumes_ShortSyntax(t *testing.T) {
	verifyVolumes(`
volumes:
  - /tmp/data:/c/data
  - sample:/c/named:ro
  - /c/target/only`,
		t,
		volume.Volume{Source: "/tmp/data", Target: "/c/data"},
		volume.Volume{Source: "sample", Target: "/c/named", Mode: "ro"},
		volume.Volume{Target: "/c/target/only"})
}

func TestVolumes_LongSyntax(t *testing.T) {
	verifyVolumes(`
volumes:
  - type: bind
    source: /tmp/data
    target: /c/data
  - type: volume
    source: sample
    target: /c/named
    read_only: true
  - type: tmpfs
    target: /c/target/only
    tmpfs:
      size: 50m`,
		t,
		volume.Volume{Source: "/tmp/data", Target: "/c/data", Type: "bind"},
		volume.Volume{Source: "sample", Target: "/c/named", Type: "volume", ReadOnly: true},
		volume.Volume{Target: "/c/target/only", Type: "tmpfs", Tmpfs: struct {
			Size string
		}{"50m"}})
}

func TestVolumes_MixedSyntax(t *testing.T) {
	verifyVolumes(`
volumes:
  - /tmp/data:/c/data
  - type: volume
    source: sample
    target: /c/named
    read_only: true
  - type: tmpfs
    target: /c/target/only
    tmpfs:
      size: 50m`,
		t,
		volume.Volume{Source: "/tmp/data", Target: "/c/data"},
		volume.Volume{Source: "sample", Target: "/c/named", Type: "volume", ReadOnly: true},
		volume.Volume{Target: "/c/target/only", Type: "tmpfs", Tmpfs: struct {
			Size string
		}{"50m"}})
}

func verifyVolumes(yamlConfig string, t *testing.T, expected ...volume.Volume) {
	actual, err := parseVolumes(yamlConfig)
	if err != nil {
		t.Fatal("Failed to parse the volumes:", err)
	}

	if len(actual) != len(expected) {
		t.Error(
			"The number of actual volumes doesn't match the expectations:",
			len(actual), "!=", len(expected))
	}

	for idx, exp := range expected {
		act := *actual[idx]

		if exp != act {
			t.Errorf("The actual volume doesn't match the expected:\n(%+v) \n(%+v)", act, exp)
		}
	}
}

func parseVolumes(yamlConfig string) ([]*volume.Volume, error) {
	var config map[string][]interface{}
	if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
		return nil, err
	}

	component := &Component{Volumes: config["volumes"]}
	return component.parseVolumes()
}
