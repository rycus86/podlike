package controller

import (
	"github.com/docker/docker/api/types"
	"github.com/rycus86/podlike/pkg/engine"
)

type Client struct {
	engine    *engine.Engine
	cgroup    string
	container *types.ContainerJSON

	closed bool
}
