package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"time"
)

type Client struct {
	api       *client.Client
	cgroup    string
	container *types.ContainerJSON

	cancelEvents context.CancelFunc

	closed bool
}

type Component struct {
	// supported options
	Image           string
	Entrypoint      interface{}
	Command         interface{}
	WorkingDir      string `yaml:"working_dir"`
	Environment     []string
	Labels          map[string]string
	Tty             bool
	StopSignal      string        `yaml:"stop_signal"`
	StopGracePeriod time.Duration `yaml:"stop_grace_period"`
	Healthcheck     *struct {
		Test        interface{}
		Interval    time.Duration
		Timeout     time.Duration
		StartPeriod time.Duration `yaml:"start_period"`
		Retries     int
	} `yaml:"healthcheck"`

	OomScoreAdj    *int  `yaml:"oom_score_adj"`
	OomKillDisable *bool `yaml:"oom_kill_disable"`

	MemoryLimit       string  `yaml:"mem_limit"`
	MemoryReservation *int64  `yaml:"mem_reservation"`
	MemorySwapLimit   string  `yaml:"memswap_limit"`
	MemorySwappiness  *int64  `yaml:"mem_swappiness"`
	ShmSize           *string `yaml:"shm_size"`

	// the parent client to the engine
	client *Client `yaml:"-"`

	// the name and container ID set in runtime
	Name      string               `yaml:"-"`
	container *types.ContainerJSON `yaml:"-"`
}

type ComponentExited struct {
	Component *Component

	StatusCode int64
	Error      error
}
