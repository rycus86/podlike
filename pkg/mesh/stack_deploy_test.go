package mesh

import (
	"encoding/json"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/podlike/pkg/component"
	"gopkg.in/yaml.v2"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestMeshStackDeploy(t *testing.T) {
	if !hasDockerCli() {
		t.Skip("Does not have access to the Docker CLI")
	}

	functionCallCount := 0

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

		if !strings.HasSuffix(request.URL.Path, "/services/create") {
			return
		}

		functionCallCount += 1

		var serviceSpec swarm.ServiceSpec
		if err := json.NewDecoder(request.Body).Decode(&serviceSpec); err != nil {
			t.Fatal("Failed to decode request body:", err)
		}

		t.Log("Stack.CreateService:", serviceSpec.Name)

		spec := serviceSpec.TaskTemplate.ContainerSpec

		if !strings.HasPrefix(spec.Image, "rycus86/podlike") {
			t.Error("Unexpected image:", spec.Image)
		}

		var comp component.Component
		if err := yaml.Unmarshal([]byte(spec.Labels["pod.component.app"]), &comp); err != nil {
			t.Fatal("Failed to unmarshal pod component:", err)
		}

		if serviceSpec.Name == "sample_web" && comp.Image != "python:2.8.x" {
			t.Error("Unexpected component image:", comp.Image)
		} else if serviceSpec.Name == "sample_db" && comp.Image != "mongo:4" {
			t.Error("Unexpected component image:", comp.Image)
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write([]byte(`{"ID": "mockSvc` + strconv.Itoa(functionCallCount) + `"}`))

	}); err != nil {
		t.Fatal("Failed to initialize mocks:", err)
	}
	defer closeMocks()

	setupFilters(mockProxy, "testdata/simple-pod.yml")

	runDockerCliCommand(
		"stack deploy",
		"-c testdata/stack.yml",
		"sample")

	if functionCallCount != 2 {
		t.Fatal("Unexpected number of calls to create service:", functionCallCount)
	}
}
