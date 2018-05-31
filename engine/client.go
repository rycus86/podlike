package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"time"
)

func (c *Client) GetComponents() ([]*Component, error) {
	var components []*Component

	for key, value := range c.container.Config.Labels {
		if strings.Index(key, "pod.component.") >= 0 {
			var component Component

			err := yaml.UnmarshalStrict([]byte(value), &component)
			if err != nil {
				return nil, err
			}

			component.init(strings.TrimPrefix(key, "pod.component."), c)

			components = append(components, &component)
		}
	}

	if composeFile, ok := c.container.Config.Labels["pod.compose.file"]; ok {
		if len(components) > 0 {
			return nil, errors.New(
				"either the individual components or a compose file should be defined, but not both")
		}

		composeContents, err := ioutil.ReadFile(composeFile)
		if err != nil {
			return nil, err
		}

		var project ComposeProject

		err = yaml.Unmarshal(composeContents, &project)
		if err != nil {
			return nil, err
		}

		for name, item := range project.Services {
			component := item
			component.init(name, c)

			components = append(components, &component)
		}
	}

	return components, nil
}

func (c *Client) Close() error {
	c.closed = true

	if c.cancelEvents != nil {
		c.cancelEvents()
	}

	return c.api.Close()
}

func NewClient() (*Client, error) {
	cgroup := getOwnCgroup()
	if cgroup == "" {
		return nil, errors.New("the application does not appear to be running in a container")
	}

	cli, err := client.NewClientWithOpts(client.WithVersion(""))
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
	cli, err = client.NewClientWithOpts(client.WithVersion(version.APIVersion))
	if err != nil {
		return nil, err
	}

	container, err := getOwnContainer(cli, cgroup)
	if err != nil {
		cli.Close()
		return nil, err
	}

	return &Client{
		api:       cli,
		cgroup:    cgroup,
		container: container,
	}, nil
}

func getOwnCgroup() string {
	contents, err := ioutil.ReadFile("/proc/self/cgroup")
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(contents), "\n") {
		parts := strings.Split(line, ":")
		if len(parts) == 3 && parts[0] == "1" {
			return parts[2]
		}
	}

	return ""
}

func getOwnContainer(c *client.Client, cgroup string) (*types.ContainerJSON, error) {
	parts := strings.Split(cgroup, "/")
	id := parts[len(parts)-1]

	if id == "" {
		return nil, errors.New("the application does not appear to be running in a container")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	container, err := c.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return &container, nil
}
