package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"github.com/rycus86/podlike/pkg/component"
	"github.com/rycus86/podlike/pkg/convert"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
	"testing"
)

var (
	nonDefaultImage = false
)

func TestTransform_Pod(t *testing.T) {
	output := Transform("testdata/stack-with-pod.yml")
	verifyTemplatedComponent(t, output, "example", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return s.Image == "rycus86/podlike:testing"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("example.container.label", "test-pod", s.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("example.label", "test-label", s.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return len(s.Ports) == 1 && s.Ports[0].Published == 8080 && s.Ports[0].Target == 4000
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/svc"
		})
}

func TestTransform_Transformer(t *testing.T) {
	output := Transform("testdata/stack-with-transformer.yml")
	verifyTemplatedComponent(t, output, "example", "transformed",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("original.label", "transformer-example", s.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("modified-original.label", "transformer-example", c.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/transform"
		})
}

func TestTransform_Templates(t *testing.T) {
	output := Transform("testdata/stack-with-templates.yml")
	verifyTemplatedComponent(t, output, "simple", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/app"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("example.label", "test-label", c.Labels)
		})

	verifyTemplatedComponent(t, output, "simple", "sidecar",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/sidecar"
		})

	verifyTemplatedComponent(t, output, "simple", "logger",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/logger"
		})
}

func TestTransform_PerServiceConfigs(t *testing.T) {
	output := Transform("testdata/stack-with-per-service-config.yml")
	verifyTemplatedComponent(t, output, "example", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return s.Image == "rycus86/podlike:testing"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/xmpl"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return sliceMatches(c.Command, "/app", "-v")
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("example.container.label", "test-pod", s.Labels)
		})

	verifyTemplatedComponent(t, output, "templated", "transformed",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/tmpl"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("modified-lbl", "label", c.Labels)
		})

	verifyTemplatedComponent(t, output, "templated", "sidecar",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/sidecar"
		})

	verifyTemplatedComponent(t, output, "templated", "logger",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/logger"
		})
}

func TestTransform_Addons(t *testing.T) {
	output := Transform("testdata/stack-with-addons.yml")
	verifyTemplatedComponent(t, output, "addons", "main",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return s.Image == "rycus86/podlike:addons"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			if len(s.Volumes) != 2 {
				return false
			}

			bind := s.Volumes[0]
			volume := s.Volumes[1]

			return bind.Source == "/var/run/docker.sock" &&
				bind.Target == "/var/run/docker.sock" &&
				bind.Type == "bind" &&
				volume.Source == "shared" &&
				volume.Target == "/var/tmp/shared" &&
				volume.Type == "volume"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/addons"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("example", "addons", c.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			if len(c.Volumes) != 1 {
				return false
			}

			volume := c.Volumes[0].(map[interface{}]interface{})
			return volume["source"] == "shared" &&
				volume["target"] == "/var/tmp/shared" &&
				volume["type"] == "volume"
		})

	verifyTemplatedComponent(t, output, "addons", "sidecar",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/sidecar"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return sliceMatches(c.Command, "app", "--serve", "--metrics", "8080")
		})

	verifyTemplatedComponent(t, output, "addons", "metrics",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/metrics"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return sliceMatches(c.Command, "--target", ":8080")
		})
}

func TestTransform_WithArgs(t *testing.T) {
	output := Transform("testdata/stack-with-args.yml")
	verifyTemplatedComponent(t, output, "with-args", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/args"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("key1", "string", c.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("key2", "42", c.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("key3", "top-level", c.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("key4-0", "str-global", c.Labels)
		})
}

func TestTransform_InlineTemplates(t *testing.T) {
	output := Transform("testdata/stack-with-inline-templates.yml")
	verifyTemplatedComponent(t, output, "inline", "main",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/inline"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return sliceMatches(c.Command, "-exec")
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("place", "controller", s.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("place", "component", c.Labels)
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("svc.name", "svc_inline", c.Labels)
		})

	verifyTemplatedComponent(t, output, "inline", "sidecar",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/sidecar"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return sliceMatches(c.Command, "--port", "8080")
		})
}

func TestTransform_Copy(t *testing.T) {
	output := Transform("testdata/stack-with-copies.yml")
	verifyTemplatedComponent(t, output, "copies", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/copy"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			var cps []string

			if copies, ok := s.Labels["pod.copy.app"]; !ok {
				return false
			} else if err := yaml.Unmarshal([]byte(copies), &cps); err != nil {
				return false
			} else {
				return sliceMatches(cps,
					"/one:/liner", "/src:/target",
					"/from:/to/a/b", "/source:/target")
			}
		})
}

func TestTransform_WithDependencies(t *testing.T) {
	output := Transform("testdata/stack-with-dependencies.yml")
	verifyTemplatedComponent(t, output, "dep", "first",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/first"
		})
	verifyTemplatedComponent(t, output, "dep", "second",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/second"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			if deps, err := c.GetDependencies(); err != nil {
				return false
			} else {
				return len(deps) == 1 &&
					deps[0].Name == "first" && deps[0].NeedsHealthyState == true
			}
		})
	verifyTemplatedComponent(t, output, "dep", "third",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/third"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			if deps, err := c.GetDependencies(); err != nil {
				return false
			} else {
				return len(deps) == 2 &&
					deps[0].Name == "first" && deps[0].NeedsHealthyState == false &&
					deps[1].Name == "second" && deps[1].NeedsHealthyState == false
			}
		})
}

func TestTransform_CustomController(t *testing.T) {
	nonDefaultImage = true
	defer func() {
		nonDefaultImage = false
	}()

	output := Transform("testdata/stack-with-custom-controller-image.yml")
	verifyTemplatedComponent(t, output, "custom-controller", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return s.Image == "forked/podlike"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/custom"
		})
}

func TestTransform_Defaults(t *testing.T) {
	output := Transform("testdata/stack-with-minimal-templates.yml")
	verifyTemplatedComponent(t, output, "minimal", "app",
		func(c *component.Component, s *types.ServiceConfig) bool {
			return strings.HasPrefix(s.Image, "rycus86/podlike:")
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return c.Image == "sample/minimal"
		},
		func(c *component.Component, s *types.ServiceConfig) bool {
			return hasLabel("sample.type", "minimal", s.Labels)
		})
}

func verifyTemplatedComponent(
	t *testing.T, output string, serviceName string, componentName string,
	expectations ...func(*component.Component, *types.ServiceConfig) bool) {

	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatal("Failed to parse the result YAML:", err)
	}

	if _, ok := result["services"]; !ok {
		t.Fatal("No services found in:", output)
	}

	services := result["services"].(map[interface{}]interface{})

	if _, ok := services[serviceName]; !ok {
		t.Fatal("Service", serviceName, "not found in:", services)
	}

	service := services[serviceName].(map[interface{}]interface{})

	if !nonDefaultImage && !strings.HasPrefix(service["image"].(string), "rycus86/podlike") {
		t.Error("Unexpected image:", service["image"])
	}

	if _, ok := service["labels"]; !ok {
		t.Fatal("No labels found in:", service)
	}

	labels := service["labels"].(map[interface{}]interface{})

	if _, ok := labels["pod.component."+componentName]; !ok {
		t.Fatal("Component", componentName, "not found in:", labels)
	}

	componentDefinition := labels["pod.component."+componentName].(string)

	convertedService, err := convertToStringKeysRecursive(service, "")
	if err != nil {
		t.Fatal("Invalid service definition for", serviceName, "in", service, err)
	}
	parsedService, err := loader.LoadService(
		serviceName, convertedService.(map[string]interface{}),
		".", os.LookupEnv)
	if err != nil {
		t.Fatal("Invalid service definition for", serviceName, "in", convertedService, err)
	}

	var comp component.Component
	err = yaml.Unmarshal([]byte(componentDefinition), &comp)
	if err != nil {
		t.Fatal("Invalid component definition in:", convertedService, err)
	}

	for idx, fn := range expectations {
		if ok := fn(&comp, parsedService); !ok {
			t.Error(fmt.Sprintf("Expectation #%d for %s failed in: %s",
				idx+1, serviceName, service))
		}
	}
}

func hasLabel(name, value string, labels interface{}) bool {
	if stringToStringMap, ok := labels.(types.Labels); ok {
		labelValue, ok := stringToStringMap[name]
		return ok && labelValue == value
	}

	converted, err := convert.ToStringToStringMap(labels)
	if err != nil {
		return false
	}

	if labelValue, ok := converted[name]; ok {
		return labelValue == value
	}

	return false
}
