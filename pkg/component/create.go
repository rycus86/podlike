package component

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/rycus86/podlike/pkg/config"
	"github.com/rycus86/podlike/pkg/convert"
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

	name := c.client.GetContainerName() + ".podlike." + c.Name

	created, err := c.engine.CreateContainer(
		containerConfig,
		hostConfig,
		name)

	if err != nil {
		if client.IsErrNotFound(err) {
			if err := c.pullImage(); err != nil {
				return "", err
			}

			created, err = c.engine.CreateContainer(
				containerConfig,
				hostConfig,
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
	entrypoint, err := convert.ToStrSlice(c.Entrypoint)
	if err != nil {
		return nil, err
	}

	command, err := convert.ToStrSlice(c.Command)
	if err != nil {
		return nil, err
	}

	envFromFiles, err := variablesFromEnvFiles(c.EnvFile)
	if err != nil {
		return nil, err
	}

	envFromVariables, err := convert.ToStringToStringMap(c.Environment)
	if err != nil {
		return nil, err
	}

	labels, err := convert.ToStringToStringMap(c.Labels)
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
		CgroupParent: c.client.GetCgroup(),

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

		Cgroup:      container.CgroupSpec("container:" + c.client.GetContainerID()),
		IpcMode:     container.IpcMode("container:" + c.client.GetContainerID()),
		NetworkMode: container.NetworkMode("container:" + c.client.GetContainerID()),

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
		tmpfs, err := convert.ToStringToStringMap(c.Tmpfs)
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
		capabilitiesToAdd, err := convert.ToStrSlice(c.CapAdd)
		if err != nil {
			return nil, err
		}

		hostConfig.CapAdd = capabilitiesToAdd
	}

	if c.CapDrop != nil {
		capabilitiesToDrop, err := convert.ToStrSlice(c.CapDrop)
		if err != nil {
			return nil, err
		}

		hostConfig.CapDrop = capabilitiesToDrop
	}

	if c.Sysctls != nil {
		sysctls, err := convert.ToStringToStringMap(c.Sysctls)
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
		hostConfig.PidMode = container.PidMode("container:" + c.client.GetContainerID())
	}

	if configuration.ShareVolumes {
		hostConfig.VolumesFrom = []string{c.client.GetContainerID()}
	}

	return &hostConfig, nil
}
