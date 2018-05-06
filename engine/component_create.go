package engine

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/rycus86/podlike/config"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

func (c *Component) createContainer(configuration *config.Configuration) (string, error) {
	if configuration.AlwaysPull {
		if err := c.pullImage(); err != nil {
			return "", err
		}
	}

	containerConfig, err := c.newContainerConfig()
	if err != nil {
		return "", err
	}

	hostConfig, err := c.newHostConfig(configuration)
	if err != nil {
		return "", err
	}

	name := c.client.container.Name + ".podlike." + c.Name

	ctxCreate, cancelCreate := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelCreate()

	created, err := c.client.api.ContainerCreate(ctxCreate,
		containerConfig,
		hostConfig,
		&network.NetworkingConfig{},
		name)

	if err != nil {
		if client.IsErrNotFound(err) {
			if err := c.pullImage(); err != nil {
				return "", err
			}

			created, err = c.client.api.ContainerCreate(ctxCreate,
				containerConfig,
				hostConfig,
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
		for _, warning := range created.Warnings {
			fmt.Sprintf("[%s] Warning: %s\n", c.Name, warning)
		}

		return created.ID, nil
	}
}

func (c *Component) newContainerConfig() (*container.Config, error) {
	entrypoint, err := asStrSlice(c.Entrypoint)
	if err != nil {
		return nil, err
	}

	command, err := asStrSlice(c.Command)
	if err != nil {
		return nil, err
	}

	envFromFiles, err := variablesFromEnvFiles(c.EnvFile)
	if err != nil {
		return nil, err
	}

	envFromVariables, err := asStringToStringMap(c.Environment)
	if err != nil {
		return nil, err
	}

	labels, err := asStringToStringMap(c.Labels)
	if err != nil {
		return nil, err
	}

	containerConfig := container.Config{
		Image:      c.Image,
		Entrypoint: entrypoint,
		Cmd:        command,
		WorkingDir: c.WorkingDir,
		Env:        mergeEnvVariables(envFromFiles, envFromVariables),
		Labels:     labels,
		OpenStdin:  c.StdinOpen,
		Tty:        c.Tty,
		StopSignal: c.StopSignal,
		User:       c.User,
	}

	if c.StopGracePeriod.Seconds() > 0 {
		stopTimeoutSeconds := int(c.StopGracePeriod.Seconds())
		containerConfig.StopTimeout = &stopTimeoutSeconds
	}

	if c.Healthcheck != nil {
		if c.Healthcheck.Disable {
			containerConfig.Healthcheck = &container.HealthConfig{
				Test: []string{"NONE"},
			}
		} else {
			testSlice, err := parseHealthcheckTest(c.Healthcheck.Test)
			if err != nil {
				return nil, err
			}

			containerConfig.Healthcheck = &container.HealthConfig{
				Test:        testSlice,
				Interval:    c.Healthcheck.Interval,
				Timeout:     c.Healthcheck.Timeout,
				StartPeriod: c.Healthcheck.StartPeriod,
				Retries:     c.Healthcheck.Retries,
			}
		}
	}

	return &containerConfig, nil
}

func (c *Component) newHostConfig(configuration *config.Configuration) (*container.HostConfig, error) {
	memLimit, err := c.getMemoryLimit()
	if err != nil {
		return nil, err
	}

	memSwapLimit, err := c.getMemorySwapLimit()
	if err != nil {
		return nil, err
	}

	resources := container.Resources{
		CgroupParent: c.client.cgroup,

		OomKillDisable: c.OomKillDisable,

		Memory:           memLimit,
		MemorySwap:       memSwapLimit,
		MemorySwappiness: c.MemorySwappiness,

		CPUShares:          c.CPUShares,
		NanoCPUs:           int64(1000000000 * c.CPUs),
		CPUPeriod:          c.CPUPeriod,
		CPUQuota:           c.CPUQuota,
		CPURealtimePeriod:  c.CPURealtimePeriod,
		CPURealtimeRuntime: c.CPURealtimeRuntime,
		CpusetCpus:         c.CpusetCpus,
		CPUCount:           c.CPUCount,
		CPUPercent:         c.CPUPercent,
	}

	if c.MemoryReservation != nil {
		resources.MemoryReservation = *c.MemoryReservation
	}

	hostConfig := container.HostConfig{
		AutoRemove: true,

		Resources: resources,

		Cgroup:      container.CgroupSpec("container:" + c.client.container.ID),
		IpcMode:     container.IpcMode("container:" + c.client.container.ID),
		NetworkMode: container.NetworkMode("container:" + c.client.container.ID),

		Privileged:     c.Privileged,
		ReadonlyRootfs: c.ReadOnly,

		OomScoreAdj: c.getOomScoreAdjust(),
	}

	if c.ShmSize != nil {
		size, err := units.RAMInBytes(*c.ShmSize)
		if err != nil {
			return nil, err
		}

		hostConfig.ShmSize = size
	}

	if c.CapAdd != nil {
		capabilitiesToAdd, err := asStrSlice(c.CapAdd)
		if err != nil {
			return nil, err
		}

		hostConfig.CapAdd = capabilitiesToAdd
	}

	if c.CapDrop != nil {
		capabilitiesToDrop, err := asStrSlice(c.CapDrop)
		if err != nil {
			return nil, err
		}

		hostConfig.CapDrop = capabilitiesToDrop
	}

	if configuration.SharePids {
		hostConfig.PidMode = container.PidMode("container:" + c.client.container.ID)
	}

	if configuration.ShareVolumes {
		hostConfig.VolumesFrom = []string{c.client.container.ID}
	}

	return &hostConfig, nil
}

func (c *Component) getOomScoreAdjust() int {
	if c.OomScoreAdj != nil {
		return *c.OomScoreAdj
	}

	score := c.client.container.HostConfig.OomScoreAdj

	if score < 1000 {
		// make it a little more likely to be killed than the controller
		return score + 1
	} else {
		return 1000
	}
}

func (c *Component) getMemoryLimit() (int64, error) {
	var (
		limit int64
		err   error
	)

	if c.MemoryLimit != "" {
		limit, err = units.RAMInBytes(c.MemoryLimit)

		if err != nil {
			return 0, err
		}
	} else {
		limit = 0
	}

	if c.client.container.HostConfig.Memory > 0 {
		if limit <= 0 || limit > c.client.container.HostConfig.Memory {
			return c.client.container.HostConfig.Memory, nil
		}
	}

	return limit, nil
}

func (c *Component) getMemorySwapLimit() (int64, error) {
	var (
		limit int64
		err   error
	)

	if c.MemorySwapLimit != "" {
		limit, err = units.RAMInBytes(c.MemorySwapLimit)

		if err != nil {
			return 0, err
		}
	} else {
		limit = 0
	}

	if c.client.container.HostConfig.MemorySwap > 0 {
		if limit <= 0 || limit > c.client.container.HostConfig.MemorySwap {
			return c.client.container.HostConfig.MemorySwap, nil
		}
	}

	return limit, nil
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

			parts := strings.Split(value, ":")
			if len(parts) != 2 {
				return errors.New(fmt.Sprintf("invalid pod.copy configuration: %s", value))
			}

			source := parts[0]
			target := parts[1]

			targetDir, targetFilename := path.Split(target)
			reader, err := createTar(source, targetFilename)
			if err != nil {
				return err
			}

			fmt.Println("Copying", source, "to", c.Name, "@", target, "...")

			return c.client.api.CopyToContainer(
				context.TODO(), c.container.ID, targetDir, reader, types.CopyToContainerOptions{})
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
