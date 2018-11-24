package mesh

import (
	"bytes"
	"encoding/json"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/podlike/pkg/component"
	"gopkg.in/yaml.v2"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestMeshServiceUpdate(t *testing.T) {
	if ! hasDockerCli() {
		t.Skip("Does not have access to the Docker CLI")
	}

	functionCalled := false

	var serviceSpec swarm.ServiceSpec
	var svcBody = new(bytes.Buffer)

	if err := initMocks(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/networks") {
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write([]byte("[]"))
			return
		}

		if strings.HasSuffix(request.URL.Path, "/services") {
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write([]byte("[]"))
			return
		}

		if strings.HasSuffix(request.URL.Path, "/services/sample-svc") {
			json.NewEncoder(svcBody).Encode(&swarm.Service{
				Spec: serviceSpec,
			})

			writer.Header().Add("Content-Length", strconv.Itoa(svcBody.Len()))
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write(svcBody.Bytes())
			return
		}

		assertHasPorts := func (published ...uint32) bool {
			for _, want := range published {
				ok := false

				for _, got := range serviceSpec.EndpointSpec.Ports {
					if want == got.PublishedPort {
						ok = true
						break
					}
				}

				if !ok {
					return false
				}
			}

			return true
		}

		if strings.HasSuffix(request.URL.Path, "/services/create") {
			if err := json.NewDecoder(request.Body).Decode(&serviceSpec); err != nil {
				t.Fatal("Failed to decode the service create request:", err)
			}

			if serviceSpec.Name != "sample-svc" {
				t.Error("Unexpected service name")
			}
			if serviceSpec.Labels["init_label"] != "start" {
				t.Error("Unexpected service labels:", serviceSpec.Labels)
			}
			if serviceSpec.Mode.Replicated == nil {
				t.Errorf("Unexpected service mode: %+v", serviceSpec.Mode)
			} else if serviceSpec.Mode.Replicated.Replicas != nil {
				t.Errorf("Unexpected replicas: %+v", serviceSpec.Mode.Replicated.Replicas)
			}

			if serviceSpec.EndpointSpec == nil {
				t.Error("No endpoint spec found")
			} else if len(serviceSpec.EndpointSpec.Ports) != 3 {
				t.Error("Unexpected port configs:", serviceSpec.EndpointSpec.Ports)
			} else if !assertHasPorts(8080, 5000, 9999) {
				t.Error("Unexpected port configs:", serviceSpec.EndpointSpec.Ports)
			}

			spec := serviceSpec.TaskTemplate.ContainerSpec

			if !strings.HasPrefix(spec.Image, "rycus86/podlike") {
				t.Error("Unexpected image:", spec.Image)
			}

			var comp component.Component
			if err := yaml.Unmarshal([]byte(spec.Labels["pod.component.app"]), &comp); err != nil {
				t.Fatal("Failed to unmarshal pod component:", err)
			}

			if comp.Image != "new/image:v1" {
				t.Error("Unexpected component image:", comp.Image)
			}

			return
		}

		if !strings.HasSuffix(request.URL.Path, "/services/update") {
			return
		}

		if err := json.NewDecoder(request.Body).Decode(&serviceSpec); err != nil {
			t.Fatal("Failed to decode the service update request:", err)
		}

		if serviceSpec.Name != "sample-svc" {
			t.Error("Unexpected service name")
		}
		if serviceSpec.Labels["init_label"] != "start" {
			t.Error("Unexpected service labels:", serviceSpec.Labels)
		} else if serviceSpec.Labels["new_label"] != "new_value" {
			t.Error("Unexpected service labels:", serviceSpec.Labels)
		}
		if serviceSpec.Mode.Replicated == nil {
			t.Errorf("Unexpected service mode: %+v", serviceSpec.Mode)
		} else if serviceSpec.Mode.Replicated.Replicas == nil {
			t.Errorf("Unexpected replicas: %+v", serviceSpec.Mode.Replicated)
		} else if *serviceSpec.Mode.Replicated.Replicas != 3 {
			t.Errorf("Unexpected replicas: %+v", serviceSpec.Mode.Replicated.Replicas)
		}

		if serviceSpec.EndpointSpec == nil {
			t.Error("No endpoint spec found")
		} else if len(serviceSpec.EndpointSpec.Ports) != 2 {
			t.Error("Unexpected port configs:", serviceSpec.EndpointSpec.Ports)
		} else if !assertHasPorts(8080, 9999) {
			t.Error("Unexpected port configs:", serviceSpec.EndpointSpec.Ports)
		}

		spec := serviceSpec.TaskTemplate.ContainerSpec

		if !strings.HasPrefix(spec.Image, "rycus86/podlike") {
			t.Error("Unexpected image:", spec.Image)
		}

		var comp component.Component
		if err := yaml.Unmarshal([]byte(spec.Labels["pod.component.app"]), &comp); err != nil {
			t.Fatal("Failed to unmarshal pod component:", err)
		}

		if comp.Image != "new/image:v2" {
			t.Error("Unexpected component image:", comp.Image)
		}

		functionCalled = true

	}); err != nil {
		t.Fatal("Failed to initialize mocks:", err)
	}
	defer closeMocks()

	mockProxy.Handle("/services/(create|update)", processServiceCreateRequests("testdata/simple-pod.yml"))

	runDockerCliCommand(
		"service create",
		"--name sample-svc",
		"--label init_label=start",
		"--publish 8080:8080",
		"--publish 5000:5000",
		"new/image:v1 run")

	runDockerCliCommand(
		"service update sample-svc",
		"--replicas 3",
		"--label-add new_label=new_value",
		"--publish-rm 5000",
		"--image new/image:v2")

	if !functionCalled {
		t.Fatal("Missing call to update service")
	}
}
