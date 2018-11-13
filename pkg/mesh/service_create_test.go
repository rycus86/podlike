package mesh

import (
	"encoding/json"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/podlike/pkg/component"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMeshSimplePod(t *testing.T) {
	if ! hasDockerCli() {
		t.Skip("Does not have access to the Docker CLI")
	}

	functionCalled := false

	if err := initMocks(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasSuffix(request.URL.Path, "/services/create") {
			return
		}

		functionCalled = true

		var serviceSpec swarm.ServiceSpec
		if err := json.NewDecoder(request.Body).Decode(&serviceSpec); err != nil {
			t.Fatal("Failed to decode request body:", err)
		}

		spec := serviceSpec.TaskTemplate.ContainerSpec

		if !strings.HasPrefix(spec.Image, "rycus86/podlike") {
			t.Error("Unexpected image:", spec.Image)
		}

		if *serviceSpec.Mode.Replicated.Replicas != 4 {
			t.Error("Unexpected number of replicas:", *serviceSpec.Mode.Replicated.Replicas)
		}
		if len(serviceSpec.EndpointSpec.Ports) != 2 {
			t.Error("Unexpected published ports:", serviceSpec.EndpointSpec.Ports)
		} else if serviceSpec.EndpointSpec.Ports[0].PublishedPort != 9999 {
			t.Error("Unexpected published ports:", serviceSpec.EndpointSpec.Ports)
		} else if serviceSpec.EndpointSpec.Ports[0].TargetPort != 7777 {
			t.Error("Unexpected published ports:", serviceSpec.EndpointSpec.Ports)
		} else if serviceSpec.EndpointSpec.Ports[1].PublishedPort != 8000 {
			t.Error("Unexpected published ports:", serviceSpec.EndpointSpec.Ports)
		} else if serviceSpec.EndpointSpec.Ports[1].TargetPort != 8080 {
			t.Error("Unexpected published ports:", serviceSpec.EndpointSpec.Ports)
		}

		if len(spec.Mounts) != 2 {
			t.Error("Unexpected number of mounts:", spec.Mounts)
		} else if spec.Mounts[0].Source != "/var/run/docker.sock" || spec.Mounts[0].Target != "/var/run/docker.sock" || spec.Mounts[0].Type != mount.TypeBind {
			t.Error("Unexpected mount:", spec.Mounts[0])
		} else if spec.Mounts[1].Source != "testvol" || spec.Mounts[1].Target != "/var/vol" || spec.Mounts[1].Type != mount.TypeVolume {
			t.Error("Unexpected mount:", spec.Mounts[1])
		}

		if spec.Hostname != "simplepod" {
			t.Error("Unexpected hostname:", spec.Hostname)
		}
		if len(spec.Hosts) != 1 {
			t.Error("Unexpected extra hosts:", spec.Hosts)
		} else if spec.Hosts[0] != "8.8.8.8 google.resolver" {
			t.Error("Unexpected extra hosts:", spec.Hosts)
		}
		if len(spec.DNSConfig.Nameservers) != 1 {
			t.Errorf("Unexpected DNS config: %+v", *spec.DNSConfig)
		} else if spec.DNSConfig.Nameservers[0] != "1.1.1.1" {
			t.Errorf("Unexpected DNS config: %+v", *spec.DNSConfig)
		}

		if spec.Labels["test.id"] != "simple-pod" {
			t.Error("Unexpected labels:", spec.Labels)
		} else if spec.Labels["svc.replicas"] != "4" {
			t.Error("Unexpected labels:", spec.Labels)
		} else if spec.Labels["pod.component.app"] == "" {
			t.Fatal("Missing pod component label:", spec.Labels)
		}

		var comp component.Component
		if err := yaml.Unmarshal([]byte(spec.Labels["pod.component.app"]), &comp); err != nil {
			t.Fatal("Failed to unmarshal pod component:", err)
		}

		if !strings.HasPrefix(comp.Image, "alpine") {
			t.Error("Unexpected image:", comp.Image)
		}
		if slice, ok := comp.Command.([]interface{}); !ok {
			t.Errorf("Unexpected command type: %T", comp.Command)
		} else if slice[0] != "sleep" || slice[1] != "60" {
			t.Error("Unexpected command:", slice)
		}
		if !comp.Tty {
			t.Error("TTY option is not set")
		}

		if env, ok := comp.Environment.(map[interface{}]interface{}); !ok {
			t.Errorf("Unexpected environment variables: %T -- %+v", comp.Environment, comp.Environment)
		} else if env["E_NIL"] != nil || env["E_EMPTY"] != "" || env["E_KEY"] != "VALUE" || env["FROM_FILE"] != "key-from-file" {
			t.Error("Unexpected environment variables:", env)
		}

		if comp.Healthcheck == nil {
			t.Fatal("Missing healthcheck configuration")
		}
		if slice, ok := comp.Healthcheck.Test.([]interface{}); !ok {
			t.Errorf("Unexpected healthcheck: %T -- %+v", comp.Healthcheck.Test, *comp.Healthcheck)
		} else if slice[0] != "CMD-SHELL" || slice[1] != "is_healthy" {
			t.Errorf("Unexpected healthcheck: %+v", *comp.Healthcheck)
		}
		if comp.Healthcheck.Interval != 10*time.Millisecond {
			t.Errorf("Unexpected healthcheck: %+v", *comp.Healthcheck)
		}
		if comp.Healthcheck.Retries != 3 {
			t.Errorf("Unexpected healthcheck: %+v", *comp.Healthcheck)
		}
		if comp.Healthcheck.StartPeriod != 2*time.Second {
			t.Errorf("Unexpected healthcheck: %+v", *comp.Healthcheck)
		}
		if comp.Healthcheck.Timeout != 3*time.Millisecond {
			t.Errorf("Unexpected healthcheck: %+v", *comp.Healthcheck)
		}

		if len(comp.Volumes) != 1 {
			t.Error("Unexpected component volumes:", comp.Volumes)
		} else if m, ok := comp.Volumes[0].(map[interface{}]interface{}); !ok {
			t.Errorf("Unexpected component volumes: %T -- %+v", comp.Volumes[0], comp.Volumes)
		} else if m["type"] != "volume" || m["source"] != "testvol" || m["target"] != "/var/vol" {
			t.Error("Unexpected component volumes:", comp.Volumes)
		}
	}); err != nil {
		t.Fatal("Failed to initialize mocks:", err)
	}
	defer closeMocks()

	mockProxy.Handle("/services/create", processServiceCreateRequests("testdata/simple-pod.yml"))

	runDockerCliCommand(
		"service create",
		"--name test-simple-pod",
		"--dns 1.1.1.1",
		"--env E_NIL",
		"--env E_EMPTY=",
		"--env E_KEY=VALUE",
		"--env-file testdata/env_file.txt",
		"--host google.resolver:8.8.8.8",
		"--hostname simplepod",
		"--health-cmd is_healthy",
		"--health-interval 10ms",
		"--health-retries 3",
		"--health-start-period 2s",
		"--health-timeout 3ms",
		"--publish 8000:8080",
		"--replicas 4",
		"--mount type=volume,source=testvol,target=/var/vol",
		"--tty --detach --no-resolve-image",
		"alpine sleep 60")

	if !functionCalled {
		t.Fatal("Missing call to create service")
	}
}
