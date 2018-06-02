package controller

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rycus86/podlike/pkg/config"
	"github.com/rycus86/podlike/pkg/convert"
	"github.com/rycus86/podlike/pkg/engine"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type verifyCreate struct {
	Verify func(name string, body map[string]interface{})
}

func TestOneComponent(t *testing.T) {
	components, err := newTestClient(map[string]string{
		"pod.component.single": "{image: sample, command: a b c}",
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

func TestComposeProject(t *testing.T) {
	components, err := newTestClient(map[string]string{
		"pod.compose.file": "testdata/docker-compose.yml",
	}, nil, nil).GetComponents()

	if err != nil {
		t.Error("Failed to get components", err)
	}

	if len(components) != 2 {
		t.Error("Unexpected number of components:", len(components))
	}

	for _, c := range components {
		if c.Name == "app" {
			if c.Image != "rycus86/demo-site" {
				t.Error("Unexpected image:", c.Image)
			}

			env, err := convert.ToStringToStringMap(c.Environment)
			if err != nil {
				t.Fatal(err)
			}

			if host, ok := env["HTTP_HOST"]; !ok || host != "127.0.0.1" {
				t.Error("Unexpected environment variables:", c.Environment)
			}

			if port, ok := env["HTTP_PORT"]; !ok || port != "12000" {
				t.Error("Unexpected environment variables:", c.Environment)
			}
		} else if c.Name == "proxy" {
			if c.Image != "nginx:1.13.10" {
				t.Error("Unexpected image:", c.Image)
			}
		} else {
			t.Error("Unexpected component name:", c.Name)
		}
	}
}

func TestStartComponent(t *testing.T) {
	labels := map[string]string{
		"pod.component.start": `
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

			return
		}

		if strings.HasSuffix(r.RequestURI, "/json") {
			w.WriteHeader(200)
			w.Write([]byte("{\"ID\": \"c0001\", \"Config\": {}}"))
			// TODO implement properly

			return
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
		engine: engine.NewEngineWithDockerClient(cli),
		container: &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:   "01234",
				Name: "mock-container",

				HostConfig: &container.HostConfig{},
			},
			Config: &container.Config{
				Labels: labels,
			},
		},
	}
}
