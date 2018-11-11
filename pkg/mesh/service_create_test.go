package mesh

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSimplePod(t *testing.T) {
	if ! hasDockerCli() {
		t.Skip("Does not have access to the Docker CLI")
	}

	initMocks(func(writer http.ResponseWriter, request *http.Request) {
		t.Log("testing:", request.URL)

		body, _ := ioutil.ReadAll(request.Body)
		t.Log("request body:\n", string(body))
	})
	defer closeMocks()

	mockProxy.Handle("/services/create", processServiceCreateRequests("testdata/simple-pod.yml"))

	runDockerCliCommand(
		"service create",
		"--name test-simple-pod",
		"--tty --detach --no-resolve-image",
		"alpine sleep 60")
}
