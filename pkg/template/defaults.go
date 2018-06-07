package template

import (
	"fmt"
	"github.com/rycus86/podlike/pkg/version"
)

func getDefaultPodTemplate() podTemplate {
	return podTemplate{
		Template: fmt.Sprintf(`
pod:
  image:   rycus86/podlike:%s
  command: -logs
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock:ro
`, getTagForPod()),
		Inline: true,
	}
}

func getTagForPod() string {
	v := version.Parse()

	if v.Tag == version.DEFAULT_VERSION {
		return "latest"
	} else {
		return v.Tag
	}
}

func getDefaultTransformerTemplate() podTemplate {
	return podTemplate{
		Template: `
app:
  image: {{ .Service.Image }}`,
		Inline: true,
	}
}
