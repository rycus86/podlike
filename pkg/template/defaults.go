package template

import (
	"fmt"
	"github.com/rycus86/podlike/pkg/version"
)

func getMinimalPodProperties(name string) map[string]interface{} {
	return map[string]interface{}{
		name: map[string]interface{}{
			"image":   "rycus86/podlike:" + getTagForPod(),
			"volumes": []string{"/var/run/docker.sock:/var/run/docker.sock:ro"},
		},
	}
}

func getDefaultPodTemplate() podTemplate {
	return podTemplate{
		Inline: fmt.Sprintf(`
pod:
  image:   rycus86/podlike:%s
  command: -logs
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock:ro
`, getTagForPod()),
	}
}

func getTagForPod() string {
	v := version.Parse()

	if v.Tag == version.DEFAULT_VERSION || v.Tag == "master" {
		return "latest"
	} else {
		return v.Tag
	}
}

func getDefaultTransformerTemplate() podTemplate {
	return podTemplate{
		Inline: `
app:
  image: {{ .Service.Image }}`,
	}
}
