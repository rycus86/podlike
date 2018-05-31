package engine

import (
	"fmt"
	"github.com/docker/docker/api/types/mount"
)

func (c *Client) GetSharedVolumeSource(volumeName string) string {
	if swarmVolume := c.getSwarmVolumeSource(volumeName); swarmVolume != "" {
		return swarmVolume
	}

	if composeVolume := c.getComposeVolumeSource(volumeName); composeVolume != "" {
		return composeVolume
	}

	return ""
}

func (c *Client) getSwarmVolumeSource(name string) string {
	for _, mnt := range c.container.HostConfig.Mounts {
		if mnt.Type != mount.TypeVolume || mnt.VolumeOptions == nil {
			continue
		}

		namespace, exists := mnt.VolumeOptions.Labels["com.docker.stack.namespace"]
		if !exists {
			continue
		}

		// TODO volumes with explicit names?
		if mnt.Source == fmt.Sprintf("%s_%s", namespace, name) {
			return mnt.Source
		}
	}

	return ""
}

func (c *Client) getComposeVolumeSource(name string) string {
	for _, mnt := range c.container.Mounts {
		projectName, exists := c.container.Config.Labels["com.docker.compose.project"]
		if !exists {
			continue
		}

		// TODO volumes with an explicit name in the Compose file won't match
		if mnt.Name == fmt.Sprintf("%s_%s", projectName, name) {
			return mnt.Name
		}
	}

	return ""
}
