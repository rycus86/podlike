package mesh

import (
	"bytes"
	"encoding/json"
	"github.com/docker/docker/api/types/swarm"
	"io/ioutil"
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
		t.Log("upd:", request.Method, request.URL)

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

		if strings.HasSuffix(request.URL.Path, "/services/create") {
			json.NewDecoder(request.Body).Decode(&serviceSpec)
			t.Logf("svc.create: %+v", serviceSpec)
			return
		}

		if !strings.HasSuffix(request.URL.Path, "/services/update") {
			return
		}

		d, _ := ioutil.ReadAll(request.Body)
		t.Log("upd.body:", string(d))

		functionCalled = true

	}); err != nil {
		t.Fatal("Failed to initialize mocks:", err)
	}
	defer closeMocks()

	// TODO should have a filter on /create as well for this test
	mockProxy.Handle("/service/.+/update", processServiceCreateRequests("testdata/simple-pod.yml"))

	mockEnableProcessLogging = true
	runDockerCliCommand(
		"service create",
		"--name sample-svc",
		"--label init_label=start",
		"new/image:v1 run")

	mockEnableProcessLogging = true
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
