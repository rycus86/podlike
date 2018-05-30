package engine

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/rycus86/podlike/config"
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
	devices, err := asDeviceMappings(c.Devices)
	if err != nil {
		return nil, err
	}

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

		Devices:           devices,
		DeviceCgroupRules: c.DeviceCgroupRules,

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

	if c.BlkioConfig != nil {
		resources.BlkioWeight = c.BlkioConfig.Weight
		resources.BlkioWeightDevice = c.BlkioConfig.WeightDevice
		resources.BlkioDeviceReadBps = c.BlkioConfig.DeviceReadBps
		resources.BlkioDeviceWriteBps = c.BlkioConfig.DeviceWriteBps
		resources.BlkioDeviceReadIOps = c.BlkioConfig.DeviceReadIOps
		resources.BlkioDeviceWriteIOps = c.BlkioConfig.DeviceWriteIOps
	}

	if c.PidsLimit != nil {
		resources.PidsLimit = *c.PidsLimit
	}

	hostConfig := container.HostConfig{
		AutoRemove: true,

		Resources: resources,

		Cgroup:      container.CgroupSpec("container:" + c.client.container.ID),
		IpcMode:     container.IpcMode("container:" + c.client.container.ID),
		NetworkMode: container.NetworkMode("container:" + c.client.container.ID),

		Privileged:     c.Privileged,
		ReadonlyRootfs: c.ReadOnly,
		Runtime:        c.Runtime,

		OomScoreAdj: c.getOomScoreAdjust(),

		GroupAdd:    c.GroupAdd,
		SecurityOpt: c.SecurityOpt,
		StorageOpt:  c.StorageOpt,
		UsernsMode:  container.UsernsMode(c.UsernsMode),

		Isolation: container.Isolation(c.Isolation),
	}

	if c.Tmpfs != nil {
		tmpfs, err := asStringToStringMap(c.Tmpfs)
		if err != nil {
			return nil, err
		}

		hostConfig.Tmpfs = tmpfs
	}

	if c.Volumes != nil {
		mounts, err := c.getMounts()
		if err != nil {
			return nil, err
		}

		hostConfig.Mounts = mounts
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

	if c.Sysctls != nil {
		sysctls, err := asStringToStringMap(c.Sysctls)
		if err != nil {
			return nil, err
		}

		hostConfig.Sysctls = sysctls
	}

	if c.Ulimits != nil {
		ulimits, err := c.getUlimits()
		if err != nil {
			return nil, err
		}

		hostConfig.Ulimits = ulimits
	}

	if c.Logging != nil {
		hostConfig.LogConfig = container.LogConfig{
			Type:   c.Logging.Driver,
			Config: c.Logging.Options,
		}
	}

	if configuration.SharePids {
		hostConfig.PidMode = container.PidMode("container:" + c.client.container.ID)
	}

	if configuration.ShareVolumes {
		hostConfig.VolumesFrom = []string{c.client.container.ID}
	}

	return &hostConfig, nil
}
