package mesh

import (
	"fmt"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"strings"
)

func convertSwarmSpecToComposeService(spec *swarm.ServiceSpec) types.ServiceConfig {
	//✓		//Name string `yaml:"-"`
	//x		//Build           BuildConfig                      `yaml:",omitempty"`
	//x		//CapAdd          []string                         `mapstructure:"cap_add" yaml:"cap_add,omitempty"`
	//x		//CapDrop         []string                         `mapstructure:"cap_drop" yaml:"cap_drop,omitempty"`
	//x		//CgroupParent    string                           `mapstructure:"cgroup_parent" yaml:"cgroup_parent,omitempty"`
	//✓		//Command         ShellCommand                     `yaml:",omitempty"`
	//✓ 	//Configs         []ServiceConfigObjConfig         `yaml:",omitempty"`
	//x		//ContainerName   string                           `mapstructure:"container_name" yaml:"container_name,omitempty"`
	//✓		//CredentialSpec  CredentialSpecConfig             `mapstructure:"credential_spec" yaml:"credential_spec,omitempty"`
	//x		//DependsOn       []string                         `mapstructure:"depends_on" yaml:"depends_on,omitempty"`
	//✓		//Deploy          DeployConfig                     `yaml:",omitempty"`
	//x		//Devices         []string                         `yaml:",omitempty"`
	//✓		//DNS             StringList                       `yaml:",omitempty"`
	//✓		//DNSSearch       StringList                       `mapstructure:"dns_search" yaml:"dns_search,omitempty"`
	//x		//DomainName      string                           `mapstructure:"domainname" yaml:"domainname,omitempty"`
	//✓		//Entrypoint      ShellCommand                     `yaml:",omitempty"`
	//✓		//Environment     MappingWithEquals                `yaml:",omitempty"`
	//x		//EnvFile         StringList                       `mapstructure:"env_file" yaml:"env_file,omitempty"`
	//x		//Expose          StringOrNumberList               `yaml:",omitempty"`
	//x		//ExternalLinks   []string                         `mapstructure:"external_links" yaml:"external_links,omitempty"`
	//✓		//ExtraHosts      HostsList                        `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty"`
	//✓		//Hostname        string                           `yaml:",omitempty"`
	//✓		//HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
	//✓		//Image           string                           `yaml:",omitempty"`
	//✓		//Init            *bool                            `yaml:",omitempty"`
	//x		//Ipc             string                           `yaml:",omitempty"`
	//✓		//Labels          Labels                           `yaml:",omitempty"`
	//x		//Links           []string                         `yaml:",omitempty"`
	//✓		//Logging         *LoggingConfig                   `yaml:",omitempty"`
	//x		//MacAddress      string                           `mapstructure:"mac_address" yaml:"mac_address,omitempty"`
	//x		//NetworkMode     string                           `mapstructure:"network_mode" yaml:"network_mode,omitempty"`
	//✓		//Networks        map[string]*ServiceNetworkConfig `yaml:",omitempty"`
	//x		//Pid             string                           `yaml:",omitempty"`
	//✓		//Ports           []ServicePortConfig              `yaml:",omitempty"`
	//x		//Privileged      bool                             `yaml:",omitempty"`
	//✓		//ReadOnly        bool                             `mapstructure:"read_only" yaml:"read_only,omitempty"`
	//x		//Restart         string                           `yaml:",omitempty"`
	//✓		//Secrets         []ServiceSecretConfig            `yaml:",omitempty"`
	//x		//SecurityOpt     []string                         `mapstructure:"security_opt" yaml:"security_opt,omitempty"`
	//✓		//StdinOpen       bool                             `mapstructure:"stdin_open" yaml:"stdin_open,omitempty"`
	//✓		//StopGracePeriod *time.Duration                   `mapstructure:"stop_grace_period" yaml:"stop_grace_period,omitempty"`
	//✓		//StopSignal      string                           `mapstructure:"stop_signal" yaml:"stop_signal,omitempty"`
	//x		//Tmpfs           StringList                       `yaml:",omitempty"`
	//✓		//Tty             bool                             `mapstructure:"tty" yaml:"tty,omitempty"`
	//x		//Ulimits         map[string]*UlimitsConfig        `yaml:",omitempty"`
	//✓		//User            string                           `yaml:",omitempty"`
	//✓		//Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
	//✓		//WorkingDir      string                           `mapstructure:"working_dir" yaml:"working_dir,omitempty"`
	//✓		//Isolation       string                           `mapstructure:"isolation" yaml:"isolation,omitempty"`

	return types.ServiceConfig{
		Name:            spec.Name,
		Command:         types.ShellCommand(spec.TaskTemplate.ContainerSpec.Args),
		Configs:         swarmConfigsToCompose(spec.TaskTemplate.ContainerSpec.Configs),
		CredentialSpec:  swarmCredentialsSpecToCompose(spec.TaskTemplate.ContainerSpec.Privileges),
		Deploy:          swarmDeployToCompose(spec),
		DNS:             types.StringList(spec.TaskTemplate.ContainerSpec.DNSConfig.Nameservers),
		DNSSearch:       types.StringList(spec.TaskTemplate.ContainerSpec.DNSConfig.Search),
		Entrypoint:      types.ShellCommand(spec.TaskTemplate.ContainerSpec.Command),
		Environment:     swarmEnvironmentToCompose(spec.TaskTemplate.ContainerSpec.Env),
		ExtraHosts:      swarmExtraHostsToCompose(spec.TaskTemplate.ContainerSpec.Hosts),
		Hostname:        spec.TaskTemplate.ContainerSpec.Hostname,
		HealthCheck:     swarmHealthCheckToCompose(spec.TaskTemplate.ContainerSpec.Healthcheck),
		Image:           spec.TaskTemplate.ContainerSpec.Image,
		Init:            spec.TaskTemplate.ContainerSpec.Init,
		Labels:          types.Labels(spec.TaskTemplate.ContainerSpec.Labels),
		Logging:         swarmLoggingToCompose(spec.TaskTemplate.LogDriver),
		Networks:        swarmNetworksToCompose(spec),
		Ports:           swarmPortsToCompose(spec.EndpointSpec.Ports),
		ReadOnly:        spec.TaskTemplate.ContainerSpec.ReadOnly,
		Secrets:         swarmSecretsToCompose(spec.TaskTemplate.ContainerSpec.Secrets),
		StdinOpen:       spec.TaskTemplate.ContainerSpec.OpenStdin,
		StopGracePeriod: spec.TaskTemplate.ContainerSpec.StopGracePeriod,
		StopSignal:      spec.TaskTemplate.ContainerSpec.StopSignal,
		Tty:             spec.TaskTemplate.ContainerSpec.TTY,
		User:            spec.TaskTemplate.ContainerSpec.User,
		Volumes:         swarmVolumeToCompose(spec.TaskTemplate.ContainerSpec.Mounts),
		WorkingDir:      spec.TaskTemplate.ContainerSpec.Dir,
		Isolation:       string(spec.TaskTemplate.ContainerSpec.Isolation),
	}
}

func swarmVolumeToCompose(mounts []mount.Mount) []types.ServiceVolumeConfig {
	var volumes []types.ServiceVolumeConfig
	for _, mnt := range mounts {
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

	return volumes
}

func swarmCredentialsSpecToCompose(privileges *swarm.Privileges) types.CredentialSpecConfig {
	var credentialSpec types.CredentialSpecConfig
	if privileges != nil {
		if privileges.CredentialSpec != nil {
			credentialSpec = types.CredentialSpecConfig{
				File:     privileges.CredentialSpec.File,
				Registry: privileges.CredentialSpec.Registry,
			}
		}
	}
	return credentialSpec
}

func swarmEnvironmentToCompose(env []string) types.MappingWithEquals {
	environment := types.MappingWithEquals{}
	for _, item := range env {
		if strings.Contains(item, "=") {
			parts := strings.SplitN(item, "=", 2)
			environment[parts[0]] = &parts[1]
		} else {
			environment[item] = nil
		}
	}

	return environment
}

func swarmExtraHostsToCompose(swarmHosts []string) types.HostsList {
	var list types.HostsList

	for _, host := range swarmHosts {
		// swarm: IP-address hostname(s)
		if v := strings.SplitN(host, " ", 2); len(v) == 2 {
			// compose: <host>:<ip>
			list = append(list, fmt.Sprintf("%s:%s", v[1], v[0]))
		}
	}

	return list
}

func swarmHealthCheckToCompose(cfg *container.HealthConfig) *types.HealthCheckConfig {
	var healthcheckConfig *types.HealthCheckConfig
	if cfg != nil {
		// check for `--no-healthcheck`
		if len(cfg.Test) == 1 && cfg.Test[0] == "NONE" {
			return &types.HealthCheckConfig{
				Disable: true,
			}
		}

		retries := uint64(cfg.Retries)
		healthcheckConfig = &types.HealthCheckConfig{
			Test:        cfg.Test,
			Timeout:     &cfg.Timeout,
			Interval:    &cfg.Interval,
			Retries:     &retries,
			StartPeriod: &cfg.StartPeriod,
		}
	}

	return healthcheckConfig
}

func swarmDeployToCompose(spec *swarm.ServiceSpec) types.DeployConfig {
	deployConfig := types.DeployConfig{
		Labels:       spec.Labels,
		Resources:    types.Resources{},
		Placement:    types.Placement{},
		EndpointMode: string(spec.EndpointSpec.Mode),
	}

	if spec.Mode.Replicated != nil {
		deployConfig.Mode = "replicated"
		deployConfig.Replicas = spec.Mode.Replicated.Replicas
	} else if spec.Mode.Global != nil {
		deployConfig.Mode = "global"
	}

	if spec.TaskTemplate.Resources != nil {
		if spec.TaskTemplate.Resources.Limits != nil {
			nanoCpus := opts.NanoCPUs(spec.TaskTemplate.Resources.Limits.NanoCPUs)

			var genericResources []types.GenericResource
			for _, r := range spec.TaskTemplate.Resources.Limits.GenericResources {
				if r.DiscreteResourceSpec != nil {
					genericResources = append(genericResources, types.GenericResource{
						DiscreteResourceSpec: &types.DiscreteGenericResource{
							Kind:  r.DiscreteResourceSpec.Kind,
							Value: r.DiscreteResourceSpec.Value,
						},
					})
				}
			}

			deployConfig.Resources.Limits = &types.Resource{
				MemoryBytes:      types.UnitBytes(spec.TaskTemplate.Resources.Limits.MemoryBytes),
				NanoCPUs:         nanoCpus.String(),
				GenericResources: genericResources,
			}
		}

		if spec.TaskTemplate.Resources.Reservations != nil {
			nanoCpus := opts.NanoCPUs(spec.TaskTemplate.Resources.Reservations.NanoCPUs)

			var genericResources []types.GenericResource
			for _, r := range spec.TaskTemplate.Resources.Reservations.GenericResources {
				if r.DiscreteResourceSpec != nil {
					genericResources = append(genericResources, types.GenericResource{
						DiscreteResourceSpec: &types.DiscreteGenericResource{
							Kind:  r.DiscreteResourceSpec.Kind,
							Value: r.DiscreteResourceSpec.Value,
						},
					})
				}
			}

			deployConfig.Resources.Reservations = &types.Resource{
				MemoryBytes:      types.UnitBytes(spec.TaskTemplate.Resources.Reservations.MemoryBytes),
				NanoCPUs:         nanoCpus.String(),
				GenericResources: genericResources,
			}
		}
	}

	if spec.TaskTemplate.Placement != nil {
		var preferences []types.PlacementPreferences
		for _, p := range spec.TaskTemplate.Placement.Preferences {
			preferences = append(preferences, types.PlacementPreferences{
				Spread: p.Spread.SpreadDescriptor,
			})
		}

		deployConfig.Placement.Constraints = spec.TaskTemplate.Placement.Constraints
		deployConfig.Placement.Preferences = preferences
	}

	if spec.TaskTemplate.RestartPolicy != nil {
		deployConfig.RestartPolicy = &types.RestartPolicy{
			Condition:   string(spec.TaskTemplate.RestartPolicy.Condition),
			MaxAttempts: spec.TaskTemplate.RestartPolicy.MaxAttempts,
			Delay:       spec.TaskTemplate.RestartPolicy.Delay,
			Window:      spec.TaskTemplate.RestartPolicy.Window,
		}
	}

	deployConfig.UpdateConfig = swarmUpdateConfigToCompose(spec.UpdateConfig)
	deployConfig.RollbackConfig = swarmUpdateConfigToCompose(spec.RollbackConfig)

	return deployConfig
}

func swarmUpdateConfigToCompose(cfg *swarm.UpdateConfig) *types.UpdateConfig {
	if cfg == nil {
		return nil
	}

	return &types.UpdateConfig{
		Parallelism:     &cfg.Parallelism,
		Delay:           cfg.Delay,
		FailureAction:   cfg.FailureAction,
		Monitor:         cfg.Monitor,
		MaxFailureRatio: cfg.MaxFailureRatio,
		Order:           cfg.Order,
	}
}

func swarmConfigsToCompose(swarmConfigs []*swarm.ConfigReference) []types.ServiceConfigObjConfig {
	var configs []types.ServiceConfigObjConfig

	for _, cfg := range swarmConfigs {
		if cfg.File != nil {
			mode := uint32(cfg.File.Mode)

			configs = append(configs, types.ServiceConfigObjConfig{
				Source: cfg.ConfigName,
				Target: cfg.File.Name,
				Mode:   &mode,
				UID:    cfg.File.UID,
				GID:    cfg.File.GID,
			})
		}
	}

	return configs
}

func swarmSecretsToCompose(swarmSecrets []*swarm.SecretReference) []types.ServiceSecretConfig {
	var secrets []types.ServiceSecretConfig

	for _, s := range swarmSecrets {
		if s.File != nil {
			mode := uint32(s.File.Mode)

			secrets = append(secrets, types.ServiceSecretConfig{
				Source: s.SecretName,
				Target: s.File.Name,
				Mode:   &mode,
				UID:    s.File.UID,
				GID:    s.File.GID,
			})
		}
	}

	return secrets
}

func swarmLoggingToCompose(logDriver *swarm.Driver) *types.LoggingConfig {
	if logDriver == nil {
		return nil
	}

	return &types.LoggingConfig{
		Driver:  logDriver.Name,
		Options: logDriver.Options,
	}
}

func swarmNetworksToCompose(spec *swarm.ServiceSpec) map[string]*types.ServiceNetworkConfig {
	networks := map[string]*types.ServiceNetworkConfig{}

	for _, n := range spec.TaskTemplate.Networks {
		networks[n.Target] = &types.ServiceNetworkConfig{
			Aliases: n.Aliases,
		}
	}

	for _, n := range spec.Networks {
		networks[n.Target] = &types.ServiceNetworkConfig{
			Aliases: n.Aliases,
		}
	}

	return networks
}

func swarmPortsToCompose(swarmPorts []swarm.PortConfig) []types.ServicePortConfig {
	var ports []types.ServicePortConfig

	for _, p := range swarmPorts {
		ports = append(ports, types.ServicePortConfig{
			Mode:      string(p.PublishMode),
			Target:    p.TargetPort,
			Published: p.PublishedPort,
			Protocol:  string(p.Protocol),
		})
	}

	return ports
}
