package mesh

import (
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/swarm"
)

func mergeComposeServiceIntoSwarmSpec(svc types.ServiceConfig, spec *swarm.ServiceSpec) {
	//Name string `yaml:"-"`
	spec.Name = svc.Name

	//Build           BuildConfig                      `yaml:",omitempty"`
	//CapAdd          []string                         `mapstructure:"cap_add" yaml:"cap_add,omitempty"`
	//CapDrop         []string                         `mapstructure:"cap_drop" yaml:"cap_drop,omitempty"`
	//CgroupParent    string                           `mapstructure:"cgroup_parent" yaml:"cgroup_parent,omitempty"`
	//Command         ShellCommand                     `yaml:",omitempty"`
	spec.TaskTemplate.ContainerSpec.Args = svc.Command

	//Configs         []ServiceConfigObjConfig         `yaml:",omitempty"`
	//ContainerName   string                           `mapstructure:"container_name" yaml:"container_name,omitempty"`
	//CredentialSpec  CredentialSpecConfig             `mapstructure:"credential_spec" yaml:"credential_spec,omitempty"`
	//DependsOn       []string                         `mapstructure:"depends_on" yaml:"depends_on,omitempty"`
	//Deploy          DeployConfig                     `yaml:",omitempty"`
	//Devices         []string                         `yaml:",omitempty"`
	//DNS             StringList                       `yaml:",omitempty"`
	//DNSSearch       StringList                       `mapstructure:"dns_search" yaml:"dns_search,omitempty"`
	//DomainName      string                           `mapstructure:"domainname" yaml:"domainname,omitempty"`
	//Entrypoint      ShellCommand                     `yaml:",omitempty"`
	spec.TaskTemplate.ContainerSpec.Command = svc.Entrypoint

	//Environment     MappingWithEquals                `yaml:",omitempty"`
	//EnvFile         StringList                       `mapstructure:"env_file" yaml:"env_file,omitempty"`
	//Expose          StringOrNumberList               `yaml:",omitempty"`
	//ExternalLinks   []string                         `mapstructure:"external_links" yaml:"external_links,omitempty"`
	//ExtraHosts      HostsList                        `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty"`
	//Hostname        string                           `yaml:",omitempty"`
	//HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
	//Image           string                           `yaml:",omitempty"`
	spec.TaskTemplate.ContainerSpec.Image = svc.Image

	//Ipc             string                           `yaml:",omitempty"`
	//Labels          Labels                           `yaml:",omitempty"`
	spec.TaskTemplate.ContainerSpec.Labels = svc.Labels

	//Links           []string                         `yaml:",omitempty"`
	//Logging         *LoggingConfig                   `yaml:",omitempty"`
	//MacAddress      string                           `mapstructure:"mac_address" yaml:"mac_address,omitempty"`
	//NetworkMode     string                           `mapstructure:"network_mode" yaml:"network_mode,omitempty"`
	//Networks        map[string]*ServiceNetworkConfig `yaml:",omitempty"`
	//Pid             string                           `yaml:",omitempty"`
	//Ports           []ServicePortConfig              `yaml:",omitempty"`
	//Privileged      bool                             `yaml:",omitempty"`
	//ReadOnly        bool                             `mapstructure:"read_only" yaml:"read_only,omitempty"`
	//Restart         string                           `yaml:",omitempty"`
	//Secrets         []ServiceSecretConfig            `yaml:",omitempty"`
	//SecurityOpt     []string                         `mapstructure:"security_opt" yaml:"security_opt,omitempty"`
	//StdinOpen       bool                             `mapstructure:"stdin_open" yaml:"stdin_open,omitempty"`
	//StopGracePeriod *time.Duration                   `mapstructure:"stop_grace_period" yaml:"stop_grace_period,omitempty"`
	//StopSignal      string                           `mapstructure:"stop_signal" yaml:"stop_signal,omitempty"`
	//Tmpfs           StringList                       `yaml:",omitempty"`
	//Tty             bool                             `mapstructure:"tty" yaml:"tty,omitempty"`
	spec.TaskTemplate.ContainerSpec.TTY = svc.Tty

	//Ulimits         map[string]*UlimitsConfig        `yaml:",omitempty"`
	//User            string                           `yaml:",omitempty"`
	//Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
	mounts, _ := convert.Volumes(svc.Volumes, map[string]types.VolumeConfig{}, convert.NewNamespace("x")) // TODO stack deploy
	spec.TaskTemplate.ContainerSpec.Mounts = mounts

	//WorkingDir      string                           `mapstructure:"working_dir" yaml:"working_dir,omitempty"`
	//Isolation       string                           `mapstructure:"isolation" yaml:"isolation,omitempty"`
}
