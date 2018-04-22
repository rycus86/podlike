package engine

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"time"
)

type Client struct {
	api       *client.Client
	container *types.ContainerJSON
}

func (c *Client) GetComponents() ([]*Component, error) {
	var components []*Component

	for key, value := range c.container.Config.Labels {
		if strings.Index(key, "pod.container.") >= 0 {
			var item Component

			err := yaml.UnmarshalStrict([]byte(value), &item)
			if err != nil {
				return nil, err
			} else {
				item.init(strings.TrimPrefix(key, "pod.container."), c)

				components = append(components, &item)
			}
		}
	}

	return components, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(""))
	if err != nil {
		return nil, err
	}

	container, err := getOwnContainer(cli)
	if err != nil {
		cli.Close()
		return nil, err
	}

	return &Client{
		api:       cli,
		container: container,
	}, nil
}

func getOwnContainer(c *client.Client) (*types.ContainerJSON, error) {
	contents, err := ioutil.ReadFile("/proc/1/cgroup")
	if err != nil {
		return nil, err
	}

	id := ""

	for _, line := range strings.Split(string(contents), "\n") {
		if strings.Contains(line, "docker/") {
			parts := strings.Split(line, "/")
			id = parts[len(parts)-1]
			break
		}
	}

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
