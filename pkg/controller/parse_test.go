package controller

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"testing"
)

func TestController_ParseInitComponents(t *testing.T) {
	cli := &Client{
		container: &types.ContainerJSON{
			Config: &container.Config{
				Labels: map[string]string{
					"pod.init.components": `
- image: sample/init-1
  environment:
    SAMPLE: test
- image: sample/init-2
  labels:
    component.type: init`,
				},
			},
			ContainerJSONBase: &types.ContainerJSONBase{
				HostConfig: &container.HostConfig{},
			},
		},
	}

	components, err := cli.GetInitComponents()
	if err != nil {
		t.Fatal("Failed to get the init components: ", err)
	}

	if len(components) != 2 {
		t.Fatal("Invalid number of init components: ", len(components))
	}

	first := components[0]

	if first.Image != "sample/init-1" {
		t.Error("Unexpected image: ", first.Image)
	}

	if env, ok := first.Environment.(map[interface{}]interface{}); !ok {
		t.Errorf("Unexpected environment variables: %+v\n", first.Environment)
	} else if env["SAMPLE"] != "test" {
		t.Error("Unexpected environment value: SAMPLE=", env["SAMPLE"])
	}

	second := components[1]

	if second.Image != "sample/init-2" {
		t.Error("Unexpected image: ", second.Image)
	}

	if labels, ok := second.Labels.(map[interface{}]interface{}); !ok {
		t.Errorf("Unexpected labels: %+v\n", second.Labels)
	} else if labels["component.type"] != "init" {
		t.Error("Unexpected label value: component.type=", labels["component.type"])
	}
}
