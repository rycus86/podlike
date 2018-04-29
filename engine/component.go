package engine

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/mattn/go-shellwords"
	"github.com/rycus86/podlike/config"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

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

	containerID, err := c.createContainer(configuration)
	if err != nil {
		return err
	}

	// save the container ID for later
	c.containerID = containerID

	if err := c.copyFilesIfNecessary(); err != nil {
		return err
	}

	if err := c.startContainer(); err != nil {
		return err
	}

	if configuration.StreamLogs {
		go c.streamLogs()
	}

	fmt.Println("Component started:", c.Name)

	return nil
}

func (c *Component) createContainer(configuration *config.Configuration) (string, error) {
	entrypoint, err := asStrSlice(c.Entrypoint)
	if err != nil {
		return "", nil
	}

	command, err := asStrSlice(c.Command)
	if err != nil {
		return "", err
	}

	if configuration.AlwaysPull {
		if err := c.pullImage(); err != nil {
			return "", err
		}
	}

	name := c.client.container.Name + ".podlike." + c.Name

	containerConfig := container.Config{
		Image:      c.Image,
		Entrypoint: entrypoint,
		Cmd:        command,
		WorkingDir: c.WorkingDir,
		Env:        c.Environment,
		Labels:     c.Labels,
		Tty:        c.Tty,
		StopSignal: c.StopSignal,
	}

	if c.StopGracePeriod.Seconds() > 0 {
		stopTimeoutSeconds := int(c.StopGracePeriod.Seconds())
		containerConfig.StopTimeout = &stopTimeoutSeconds
	}

	hostConfig := container.HostConfig{
		AutoRemove: true,

		Cgroup:      container.CgroupSpec(c.client.cgroup),
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
		&containerConfig,
		&hostConfig,
		&network.NetworkingConfig{},
		name)

	if err != nil {
		if client.IsErrNotFound(err) {
			if err := c.pullImage(); err != nil {
				return "", err
			}

			created, err = c.client.api.ContainerCreate(ctxCreate,
				&containerConfig,
				&hostConfig,
				&network.NetworkingConfig{},
				name)

			if err != nil {
				return "", err
			} else {
				return created.ID, nil
			}
		} else {
			return "", err
		}
	} else {
		// TODO handle warnings

		return created.ID, nil
	}
}

func asStrSlice(value interface{}) (strslice.StrSlice, error) {
	if value == nil {
		return nil, nil
	}

	stringValue, ok := value.(string)
	if ok {
		return shellwords.Parse(stringValue)
	}

	sliceValue, ok := value.([]string)
	if ok {
		return sliceValue, nil
	} else {
		return nil, errors.New(fmt.Sprintf("invalid string or slice: %T %+v", value, value))
	}
}

func (c *Component) pullImage() error {
	fmt.Println("Pulling image:", c.Image)

	// TODO is context.Background() appropriate here?
	if reader, err := c.client.api.ImagePull(context.Background(), c.Image, types.ImagePullOptions{}); err != nil {
		return err
	} else {
		defer reader.Close()

		ioutil.ReadAll(reader)

		return nil
	}
}

func (c *Component) copyFilesIfNecessary() error {
	for key, value := range c.client.container.Config.Labels {
		if strings.Index(key, "pod.copy.") >= 0 {
			if target := strings.TrimPrefix(key, "pod.copy."); target != c.Name {
				continue
			}

			// TODO error handling here
			source := strings.Split(value, ":")[0]
			target := strings.Split(value, ":")[1]

			targetDir, targetFilename := path.Split(target)
			reader, err := createTar(source, targetFilename)
			if err != nil {
				return err
			}

			fmt.Println("Copying", source, "to", c.Name, "@", target, "...")

			return c.client.api.CopyToContainer(
				context.TODO(), c.containerID, targetDir, reader, types.CopyToContainerOptions{})
		}
	}

	return nil
}

func createTar(path, filename string) (io.Reader, error) {
	var b bytes.Buffer

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tw := tar.NewWriter(&b)
	hdr := tar.Header{
		Name: filename,
		Mode: 0644,
		Size: fi.Size(),
	}
	if err := tw.WriteHeader(&hdr); err != nil {
		return nil, err
	}

	if _, err = tw.Write(contents); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	return &b, nil
}

func (c *Component) streamLogs() {
	if reader, err := c.client.api.ContainerLogs(context.Background(), c.containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}); err == nil {
		defer reader.Close()

		fmt.Println("Streaming logs for", c.Name)

		bufReader := bufio.NewReader(reader)
		defer reader.Close()

		for {
			out, _, err := bufReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					fmt.Println("Stopped streaming logs for", c.Name, ":", err)
				}
				return
			}

			streamType := "out"
			if out[0] == 2 {
				streamType = "err"
			}

			fmt.Printf("[%s] %s: %s\n", streamType, c.Name, strings.TrimSpace(string(out[8:])))
		}
	}
}

func (c *Component) startContainer() error {
	ctxStart, cancelStart := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStart()

	return c.client.api.ContainerStart(ctxStart, c.containerID, types.ContainerStartOptions{})
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
