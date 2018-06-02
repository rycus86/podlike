package api

import (
	"github.com/docker/docker/api/types/container"
)

type Controller interface {
	GetContainerID() string
	GetContainerName() string
	GetCgroup() string
	GetLabels() map[string]string
	GetHostConfig() *container.HostConfig
	GetSharedVolumeSource(source string) string
}
