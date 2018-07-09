package controller

import (
	"errors"
	"fmt"
	dtc "github.com/docker/docker/api/types/container"
	"github.com/rycus86/podlike/pkg/component"
	"github.com/rycus86/podlike/pkg/engine"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

func (c *Client) GetInitComponents() ([]*component.Component, error) {
	var components []*component.Component

	if initConfigs, ok := c.container.Config.Labels["pod.init.components"]; ok {
		err := yaml.UnmarshalStrict([]byte(initConfigs), &components)
		if err != nil {
			return nil, err
		}

		for idx, comp := range components {
			comp.DisableHealthChecking()
			comp.Initialize(fmt.Sprintf("init-%d", idx+1), c, c.engine)
		}
	}

	return components, nil
}

func (c *Client) GetComponents() ([]*component.Component, error) {
	var components []*component.Component

	for key, value := range c.container.Config.Labels {
		if strings.HasPrefix(key, "pod.component.") {
			var comp component.Component

			err := yaml.UnmarshalStrict([]byte(value), &comp)
			if err != nil {
				return nil, err
			}

			comp.Initialize(strings.TrimPrefix(key, "pod.component."), c, c.engine)

			components = append(components, &comp)
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

		var project component.ComposeProject

		err = yaml.Unmarshal(composeContents, &project)
		if err != nil {
			return nil, err
		}

		for name, item := range project.Services {
			comp := item
			comp.Initialize(name, c, c.engine)

			components = append(components, &comp)
		}
	}

	return components, nil
}

func (c *Client) Close() error {
	c.closed = true

	return c.engine.Close()
}

func (c *Client) GetContainerID() string {
	return c.container.ID
}

func (c *Client) GetContainerName() string {
	return c.container.Name
}

func (c *Client) GetCgroup() string {
	return c.cgroup
}

func (c *Client) GetLabels() map[string]string {
	return c.container.Config.Labels
}

func (c *Client) GetHostConfig() *dtc.HostConfig {
	return c.container.HostConfig
}

func NewClient() (*Client, error) {
	cgroup := getOwnCgroup()
	containerID := getOwnContainerID()

	if cgroup == "" || containerID == "" {
		return nil, errors.New("the application does not appear to be running in a container")
	}

	eng, err := engine.NewEngine()
	if err != nil {
		return nil, err
	}

	c := &Client{
		engine: eng,
		cgroup: cgroup,
	}

	container, err := eng.InspectContainer(containerID)
	if err != nil {
		eng.Close()
		return nil, err
	}
	c.container = container

	return c, nil
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

func getOwnContainerID() string {
	parts := strings.Split(getOwnCgroup(), "/")
	return parts[len(parts)-1]
}
