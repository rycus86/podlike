package mesh

import (
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/swarm"
	"strings"
)

func convertSwarmSpecToComposeService(spec *swarm.ServiceSpec) types.ServiceConfig {
	var volumes []types.ServiceVolumeConfig
	for _, mnt := range spec.TaskTemplate.ContainerSpec.Mounts {
		var (
			bindOptions   *types.ServiceVolumeBind
			volumeOptions *types.ServiceVolumeVolume
			tmpfsOptions  *types.ServiceVolumeTmpfs
		)

		if mnt.BindOptions != nil {
			bindOptions = &types.ServiceVolumeBind{
				Propagation: string(mnt.BindOptions.Propagation),
			}
		}
		if mnt.VolumeOptions != nil {
			volumeOptions = &types.ServiceVolumeVolume{
				NoCopy: mnt.VolumeOptions.NoCopy,
			}
		}
		if mnt.TmpfsOptions != nil {
			tmpfsOptions = &types.ServiceVolumeTmpfs{
				Size: mnt.TmpfsOptions.SizeBytes,
			}
		}

		volumes = append(volumes, types.ServiceVolumeConfig{
			Type:        string(mnt.Type),
			Source:      mnt.Source,
			Target:      mnt.Target,
			ReadOnly:    mnt.ReadOnly,
			Consistency: string(mnt.Consistency),

			Bind:   bindOptions,
			Volume: volumeOptions,
			Tmpfs:  tmpfsOptions,
		})
	}

	var credentialSpec types.CredentialSpecConfig
	if spec.TaskTemplate.ContainerSpec.Privileges != nil {
		if spec.TaskTemplate.ContainerSpec.Privileges.CredentialSpec != nil {
			credentialSpec = types.CredentialSpecConfig{
				File:     spec.TaskTemplate.ContainerSpec.Privileges.CredentialSpec.File,
				Registry: spec.TaskTemplate.ContainerSpec.Privileges.CredentialSpec.Registry,
			}
		}
	}

	environment := types.MappingWithEquals{}
	for _, item := range spec.TaskTemplate.ContainerSpec.Env {
		if strings.Contains(item, "=") {
			parts := strings.SplitN(item, "=", 2)
			environment[parts[0]] = &parts[1]
		} else {
			environment[item] = nil
		}
	}

	var healthcheckConfig *types.HealthCheckConfig
	if spec.TaskTemplate.ContainerSpec.Healthcheck != nil {
		retries := uint64(spec.TaskTemplate.ContainerSpec.Healthcheck.Retries)
		healthcheckConfig = &types.HealthCheckConfig{
			Test:        spec.TaskTemplate.ContainerSpec.Healthcheck.Test,
			Timeout:     &spec.TaskTemplate.ContainerSpec.Healthcheck.Timeout,
			Interval:    &spec.TaskTemplate.ContainerSpec.Healthcheck.Interval,
			Retries:     &retries,
			StartPeriod: &spec.TaskTemplate.ContainerSpec.Healthcheck.StartPeriod,
		}
	}

	// TODO
	return types.ServiceConfig{
//✓		//Name string `yaml:"-"`
		Name: spec.Name,

//x		//Build           BuildConfig                      `yaml:",omitempty"`
//x		//CapAdd          []string                         `mapstructure:"cap_add" yaml:"cap_add,omitempty"`
//x		//CapDrop         []string                         `mapstructure:"cap_drop" yaml:"cap_drop,omitempty"`
//x		//CgroupParent    string                           `mapstructure:"cgroup_parent" yaml:"cgroup_parent,omitempty"`
//✓		//Command         ShellCommand                     `yaml:",omitempty"`
		Command: types.ShellCommand(spec.TaskTemplate.ContainerSpec.Args),

		//Configs         []ServiceConfigObjConfig         `yaml:",omitempty"`
//x		//ContainerName   string                           `mapstructure:"container_name" yaml:"container_name,omitempty"`
//✓		//CredentialSpec  CredentialSpecConfig             `mapstructure:"credential_spec" yaml:"credential_spec,omitempty"`
		CredentialSpec: credentialSpec,
//x		//DependsOn       []string                         `mapstructure:"depends_on" yaml:"depends_on,omitempty"`
		//Deploy          DeployConfig                     `yaml:",omitempty"`
//x		//Devices         []string                         `yaml:",omitempty"`
//p		//DNS             StringList                       `yaml:",omitempty"`
//p		//DNSSearch       StringList                       `mapstructure:"dns_search" yaml:"dns_search,omitempty"`
//p		//DomainName      string                           `mapstructure:"domainname" yaml:"domainname,omitempty"`
//✓		//Entrypoint      ShellCommand                     `yaml:",omitempty"`
		Entrypoint: types.ShellCommand(spec.TaskTemplate.ContainerSpec.Command),

//✓		//Environment     MappingWithEquals                `yaml:",omitempty"`
		Environment: environment,
//?		//EnvFile         StringList                       `mapstructure:"env_file" yaml:"env_file,omitempty"`
//x		//Expose          StringOrNumberList               `yaml:",omitempty"`
//x		//ExternalLinks   []string                         `mapstructure:"external_links" yaml:"external_links,omitempty"`
//p		//ExtraHosts      HostsList                        `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty"`
//p		//Hostname        string                           `yaml:",omitempty"`
//✓		//HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
		HealthCheck: healthcheckConfig,
//✓		//Image           string                           `yaml:",omitempty"`
		Image: spec.TaskTemplate.ContainerSpec.Image,

//p		//Ipc             string                           `yaml:",omitempty"`
//✓		//Labels          Labels                           `yaml:",omitempty"`
		Labels: types.Labels(spec.TaskTemplate.ContainerSpec.Labels),

//x		//Links           []string                         `yaml:",omitempty"`
		//Logging         *LoggingConfig                   `yaml:",omitempty"`
//p		//MacAddress      string                           `mapstructure:"mac_address" yaml:"mac_address,omitempty"`
//x		//NetworkMode     string                           `mapstructure:"network_mode" yaml:"network_mode,omitempty"`
//p		//Networks        map[string]*ServiceNetworkConfig `yaml:",omitempty"`
//p		//Pid             string                           `yaml:",omitempty"`
//p		//Ports           []ServicePortConfig              `yaml:",omitempty"`
//?		//Privileged      bool                             `yaml:",omitempty"`

//✓		//ReadOnly        bool                             `mapstructure:"read_only" yaml:"read_only,omitempty"`
		ReadOnly: spec.TaskTemplate.ContainerSpec.ReadOnly,
//x		//Restart         string                           `yaml:",omitempty"`
		//Secrets         []ServiceSecretConfig            `yaml:",omitempty"`
//x		//SecurityOpt     []string                         `mapstructure:"security_opt" yaml:"security_opt,omitempty"`
//✓		//StdinOpen       bool                             `mapstructure:"stdin_open" yaml:"stdin_open,omitempty"`
		StdinOpen: spec.TaskTemplate.ContainerSpec.OpenStdin,
//✓		//StopGracePeriod *time.Duration                   `mapstructure:"stop_grace_period" yaml:"stop_grace_period,omitempty"`
		StopGracePeriod: spec.TaskTemplate.ContainerSpec.StopGracePeriod,
//✓		//StopSignal      string                           `mapstructure:"stop_signal" yaml:"stop_signal,omitempty"`
		StopSignal: spec.TaskTemplate.ContainerSpec.StopSignal,
		//Tmpfs           StringList                       `yaml:",omitempty"`
//✓		//Tty             bool                             `mapstructure:"tty" yaml:"tty,omitempty"`
		Tty: spec.TaskTemplate.ContainerSpec.TTY,

//?		//Ulimits         map[string]*UlimitsConfig        `yaml:",omitempty"`
//✓		//User            string                           `yaml:",omitempty"`
		User: spec.TaskTemplate.ContainerSpec.User,
//✓		//Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
		Volumes: volumes,

//✓		//WorkingDir      string                           `mapstructure:"working_dir" yaml:"working_dir,omitempty"`
		WorkingDir: spec.TaskTemplate.ContainerSpec.Dir,
//✓		//Isolation       string                           `mapstructure:"isolation" yaml:"isolation,omitempty"`
		Isolation: string(spec.TaskTemplate.ContainerSpec.Isolation),
	}
}

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
