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
	cgroup := getOwnCgroup()
	if cgroup == "" {
		return nil, errors.New("the application does not appear to be running in a container")
	}

	cli, err := client.NewClientWithOpts(client.WithVersion(""))
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
		if len(parts) == 3 {
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
