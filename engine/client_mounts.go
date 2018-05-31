package engine

import (
	"context"
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
	if _, isInSwarm := c.container.Config.Labels["com.docker.stack.namespace"]; !isInSwarm {
		return ""
	}

	for _, mnt := range c.container.HostConfig.Mounts {
		if mnt.Type != mount.TypeVolume || mnt.VolumeOptions == nil {
			continue
		}

		explicitName, exists := mnt.VolumeOptions.Labels["com.github.rycus86.volume-ref"]
		if exists && explicitName == name {
			return mnt.Source
		}

		namespace, exists := mnt.VolumeOptions.Labels["com.docker.stack.namespace"]
		if !exists {
			continue
		}

		if mnt.Source == fmt.Sprintf("%s_%s", namespace, name) {
			return mnt.Source
		}
	}

	return ""
}

func (c *Client) getComposeVolumeSource(name string) string {
	projectName, exists := c.container.Config.Labels["com.docker.compose.project"]
	if !exists {
		return ""
	}

	for _, mnt := range c.container.Mounts {
		if mnt.Type != mount.TypeVolume {
			continue
		}

		// TODO this could be cached
		// TODO context with timeout
		volume, err := c.api.VolumeInspect(context.TODO(), mnt.Name)
		if err != nil {
			fmt.Println("Failed to get volume information for", mnt.Name)
			continue
		}

		explicitName, exists := volume.Labels["com.github.rycus86.volume-ref"]
		if exists && explicitName == name {
			return mnt.Name
		}

		if mnt.Name == fmt.Sprintf("%s_%s", projectName, name) {
			return mnt.Name
		}
	}

	return ""
}
