package api

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"io"
	"time"
)

type Engine interface {
	InspectContainer(containerID string) (*types.ContainerJSON, error)
	CreateContainer(containerConfig *container.Config, hostConfig *container.HostConfig, name string) (container.ContainerCreateCreatedBody, error)
	StartContainer(containerID string) error
	StopContainer(containerID string, timeout *time.Duration) error
	RemoveContainer(containerID string) error
	CopyToContainer(containerID string, destPath string, content io.Reader) error
	WaitContainer(containerID string) (<-chan container.ContainerWaitOKBody, <-chan error)
	StreamLogs(containerID string) (io.ReadCloser, error)
	PullImage(reference string) (io.ReadCloser, error)
	InspectVolume(name string) (types.Volume, error)
	WatchHealthcheckEvents() (<-chan events.Message, <-chan error)
}
