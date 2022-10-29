package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/rycus86/podlike/pkg/config"
)

func getRegistryAuth(filename string) *config.RegistryAuth {
	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		return &config.RegistryAuth{}
	}

	var auth config.RegistryAuth

	if err := json.Unmarshal(contents, &auth); err != nil {
		fmt.Println("[warning] Could not unmarshal auth file: ", err)
		return &config.RegistryAuth{}
	}

	return &auth
}

func NewEngineWithDockerClient(client *docker.Client) *Engine {
	return &Engine{
		api:  client,
		auth: getRegistryAuth("/var/run/secrets/podlike/dockerregistryauth.json"),
	}
}

func NewEngine() (*Engine, error) {
	cli, err := newDockerClient()
	if err != nil {
		return nil, err
	}

	return NewEngineWithDockerClient(cli), nil
}

func newDockerClient() (*docker.Client, error) {
	cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithVersion(""))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	version, err := cli.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println("Using API version:", version.APIVersion)

	// close
	cli.Close()
	// and reopen with the actual API version
	cli, err = docker.NewClientWithOpts(docker.FromEnv, docker.WithVersion(version.APIVersion))
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (e *Engine) Close() error {
	if e.cancelEvents != nil {
		e.cancelEvents()
	}

	return e.api.Close()
}
