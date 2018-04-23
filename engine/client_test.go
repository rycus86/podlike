package engine

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/rycus86/podlike/config"
)

type verifyCreate struct {
	Verify func(name string, body map[string]interface{})
}

func TestOneComponent(t *testing.T) {
	components, err := newTestClient(map[string]string{
		"pod.container.single": "{image: sample, command: a b c}",
	}, nil, nil).GetComponents()

	if err != nil {
		t.Error("Failed to get components", err)
	}

	if len(components) != 1 {
		t.Error("Unexpected number of components:", len(components))
	}

	comp := components[0]

	if comp.Name != "single" {
		t.Error("Invalid name:", comp.Name)
	}

	if comp.Image != "sample" {
		t.Error("Invalid image:", comp.Image)
	}

	if comp.Command != "a b c" {
		t.Error("Invalid command:", comp.Command)
	}
}

func TestStartComponent(t *testing.T) {
	labels := map[string]string{
		"pod.container.start": `
image: sample
command: echo test
`}
	verifier := &verifyCreate{
		Verify: func(name string, body map[string]interface{}) {
			if name != "mock-container.podlike.start" {
				t.Error("Invalid name requested:", name)
			}

			if body["Cmd"].([]interface{})[0] != "echo" ||
				body["Cmd"].([]interface{})[1] != "test" {

				t.Error("Invalid command requested:", body["Cmd"])
			}
		},
	}
	created := &container.ContainerCreateCreatedBody{ID: "c0001"}

	cli := newTestClient(labels, verifier, created)

	components, err := cli.GetComponents()

	if err != nil || len(components) != 1 {
		t.Error("Failed to get components", err)
	}

	err = components[0].Start(&config.Configuration{})
	if err != nil {
		t.Error(err)
	}
}

func newTestClient(
	labels map[string]string,
	createVerifier *verifyCreate, createResponse *container.ContainerCreateCreatedBody) *Client {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request URI:", r.RequestURI)

		if strings.Contains(r.RequestURI, "/containers/create") {
			name := r.URL.Query().Get("name")

			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)

			if createVerifier != nil {
				(*createVerifier).Verify(name, body)
			}

			response, _ := json.Marshal(createResponse)

			w.WriteHeader(200)
			w.Write(response)
		}
	}))

	cli, err := client.NewClientWithOpts(
		client.WithHTTPClient(server.Client()),
		client.WithHost(server.URL),
	)

	if err != nil {
		panic(err)
	}

	return &Client{
		api: cli,
		container: &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:   "01234",
				Name: "mock-container",
			},
			Config: &container.Config{
				Labels: labels,
			},
		},
	}
}
