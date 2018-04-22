package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/mattn/go-shellwords"
	"github.com/rycus86/podlike/config"
	"io/ioutil"
	"time"
)

type Component struct {
	client *Client `yaml:"-"`

	Image   string
	Command string `yaml:"command,omitempty"`

	Name        string `yaml:"-"`
	containerID string `yaml:"-"`
}

type ComponentExited struct {
	Component *Component

	StatusCode int64
	Error      error
}

func (c *Component) init(name string, client *Client) {
	c.Name = name
	c.client = client

	c.warnForSettings()
}

func (c *Component) warnForSettings() {
	// TODO :>
	// warning about memory limit issues
	// warning about logging driver
	// warning about stop grace period -- not visible on the container in Swarm
	// swarm service labels are not visible
	// the daemon socket will always have to be shared if it comes from a mount
	// secrets/configs ?
}

func (c *Component) Start(configuration *config.Configuration) error {
	fmt.Println("Starting component:", c.Name)

	parsedCommand, err := shellwords.Parse(c.Command)
	if err != nil {
		return nil
	}

	hostConfig := container.HostConfig{
		AutoRemove: true,

		Cgroup:      container.CgroupSpec("container:" + c.client.container.ID),
		IpcMode:     container.IpcMode("container:" + c.client.container.ID),
		NetworkMode: container.NetworkMode("container:" + c.client.container.ID),
	}

	if configuration.SharePids {
		hostConfig.PidMode = container.PidMode("container:" + c.client.container.ID)
	}

	if configuration.ShareVolumes {
		hostConfig.VolumesFrom = []string{c.client.container.ID}
	}

	ctxCreate, cancelCreate := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelCreate()

	created, err := c.client.api.ContainerCreate(ctxCreate,
		&container.Config{
			Image: c.Image,
			Cmd:   strslice.StrSlice(parsedCommand),
		},
		&hostConfig,
		&network.NetworkingConfig{},
		c.client.container.Name+".podlike."+c.Name)

	if err != nil {
		return err
	}

	c.containerID = created.ID

	ctxStart, cancelStart := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStart()

	if err = c.client.api.ContainerStart(ctxStart, created.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	fmt.Println("Component started:", c.Name)

	return nil
}

func (c *Component) Stop() error {
	fmt.Println("Stopping container:", c.Name)

	if c.containerID == "" {
		return errors.New("Container is not running for component: " + c.Name)
	}

	ctxStop, cancelStop := context.WithTimeout(context.Background(), 15*time.Second) // TODO stop grace period + extra
	defer cancelStop()

	err := c.client.api.ContainerStop(ctxStop, c.containerID, nil) // TODO stop grace period here or on create?
	if err != nil {
		fmt.Println("Failed to stop the container:", err)
		// TODO what to do here?
	}

	c.printLogs()  // TODO get logs -- remove this

	ctxRemove, cancelRemove := context.WithTimeout(context.Background(), 15*time.Second) // TODO should still fit in the wrapper grace period
	defer cancelRemove()

	err = c.client.api.ContainerRemove(ctxRemove, c.containerID, types.ContainerRemoveOptions{
		Force: true,
	})

	if err != nil {
		fmt.Println("Failed to remove the container:", err)
	}

	return err
}

func (c *Component) printLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if logs, err := c.client.api.ContainerLogs(ctx, c.containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	}); err == nil {
		defer logs.Close()

		contents, _ := ioutil.ReadAll(logs)
		fmt.Println("Logs:", string(contents))
	}
}

func (c *Component) WaitFor(exitChan chan<- ComponentExited) {
	if c.containerID == "" {
		exitChan <- ComponentExited{
			Component: c,
			Error:     errors.New("component not started"),
		}
		return
	}

	waitChan, errChan := c.client.api.ContainerWait(context.Background(), c.containerID, container.WaitConditionNotRunning)

	for {
		select {
		case exit := <-waitChan:
			if exit.Error != nil {
				exitChan <- ComponentExited{
					Component: c,
					Error:     errors.New(exit.Error.Message),
				}
			} else {
				exitChan <- ComponentExited{
					Component:  c,
					StatusCode: exit.StatusCode,
				}
			}

		case err := <-errChan:
			exitChan <- ComponentExited{
				Component: c,
				Error:     err,
			}
		}
	}
}
