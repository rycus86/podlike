package controller

import (
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"strings"
	"sync"
)

var (
	volumeInitializer sync.Once
	volumeMappings    map[string]string
)

func (c *Client) GetSharedVolumeSource(source string) string {
	volumeInitializer.Do(c.ensureVolumesAreMapped)

	return volumeMappings[source]
}

func (c *Client) ensureVolumesAreMapped() {
	volumeMappings = map[string]string{}

	for _, mnt := range c.container.Mounts {
		if mnt.Type != mount.TypeVolume {
			continue
		}

		norm := c.fetchNormalizedVolumeName(mnt.Name)
		if norm != "" {
			volumeMappings[norm] = mnt.Name
		}
	}
}

func (c *Client) fetchNormalizedVolumeName(name string) string {
	volume, err := c.engine.InspectVolume(name)
	if err != nil {
		fmt.Println("Failed to get volume information for", name)
		return ""
	}

	if explicitName, ok := volume.Labels["com.github.rycus86.podlike.volume-ref"]; ok {
		return explicitName
	}

	if swarmNamespace, ok := volume.Labels["com.docker.stack.namespace"]; ok {
		if strings.HasPrefix(volume.Name, swarmNamespace+"_") {
			return strings.TrimPrefix(volume.Name, swarmNamespace+"_")
		}
	}

	if composeName, ok := volume.Labels["com.docker.compose.volume"]; ok {
		return composeName
	}

	return ""
}
