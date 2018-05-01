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
	Image           string
	Entrypoint      interface{}
	Command         interface{}
	WorkingDir      string            `yaml:"working_dir"`
	Environment     []string          // TODO this can be a map too, plus `env_file` support
	Labels          map[string]string // TODO this can be a list of KEY=VALUE too
	Privileged      bool
	ReadOnly        bool `yaml:"read_only"`
	StdinOpen       bool `yaml:"stdin_open"`
	Tty             bool
	StopSignal      string        `yaml:"stop_signal"`
	StopGracePeriod time.Duration `yaml:"stop_grace_period"`
	User            string

	Healthcheck *Healthcheck

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

type Healthcheck struct {
	Test        interface{}
	Interval    time.Duration
	Timeout     time.Duration
	StartPeriod time.Duration `yaml:"start_period"`
	Retries     int
	Disable     bool
}

type ComponentExited struct {
	Component *Component

	StatusCode int64
	Error      error
}
