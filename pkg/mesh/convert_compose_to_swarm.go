package mesh

import (
	"fmt"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"os"
	"strings"
)

func mergeComposeServiceIntoSwarmSpec(svc *types.ServiceConfig, spec *swarm.ServiceSpec) {
	//✓		//Name string `yaml:"-"`
	//x		//Build           BuildConfig                      `yaml:",omitempty"`
	//x		//CapAdd          []string                         `mapstructure:"cap_add" yaml:"cap_add,omitempty"`
	//x		//CapDrop         []string                         `mapstructure:"cap_drop" yaml:"cap_drop,omitempty"`
	//x		//CgroupParent    string                           `mapstructure:"cgroup_parent" yaml:"cgroup_parent,omitempty"`
	//✓		//Command         ShellCommand                     `yaml:",omitempty"`
	//✓		//Configs         []ServiceConfigObjConfig         `yaml:",omitempty"`
	//x		//ContainerName   string                           `mapstructure:"container_name" yaml:"container_name,omitempty"`
	//TODO 	//CredentialSpec  CredentialSpecConfig             `mapstructure:"credential_spec" yaml:"credential_spec,omitempty"`
	//x		//DependsOn       []string                         `mapstructure:"depends_on" yaml:"depends_on,omitempty"`
	//TODO 	//Deploy          DeployConfig                     `yaml:",omitempty"`
	//x		//Devices         []string                         `yaml:",omitempty"`
	//✓		//DNS             StringList                       `yaml:",omitempty"`
	//✓		//DNSSearch       StringList                       `mapstructure:"dns_search" yaml:"dns_search,omitempty"`
	//x		//DomainName      string                           `mapstructure:"domainname" yaml:"domainname,omitempty"`
	//✓		//Entrypoint      ShellCommand                     `yaml:",omitempty"`
	//TODO 	//Environment     MappingWithEquals                `yaml:",omitempty"`
	//x		//EnvFile         StringList                       `mapstructure:"env_file" yaml:"env_file,omitempty"`
	//x		//Expose          StringOrNumberList               `yaml:",omitempty"`
	//x		//ExternalLinks   []string                         `mapstructure:"external_links" yaml:"external_links,omitempty"`
	//✓		//ExtraHosts      HostsList                        `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty"`
	//✓		//Hostname        string                           `yaml:",omitempty"`
	//TODO 	//HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
	//✓		//Image           string                           `yaml:",omitempty"`
	//✓		//Init            *bool                            `yaml:",omitempty"`
	//x		//Ipc             string                           `yaml:",omitempty"`
	//✓		//Labels          Labels                           `yaml:",omitempty"`
	//x		//Links           []string                         `yaml:",omitempty"`
	//TODO 	//Logging         *LoggingConfig                   `yaml:",omitempty"`
	//x		//MacAddress      string                           `mapstructure:"mac_address" yaml:"mac_address,omitempty"`
	//x		//NetworkMode     string                           `mapstructure:"network_mode" yaml:"network_mode,omitempty"`
	//TODO 	//Networks        map[string]*ServiceNetworkConfig `yaml:",omitempty"`
	//x		//Pid             string                           `yaml:",omitempty"`
	//TODO 	//Ports           []ServicePortConfig              `yaml:",omitempty"`
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
	//TODO 	//Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
	//✓		//WorkingDir      string                           `mapstructure:"working_dir" yaml:"working_dir,omitempty"`
	//✓		//Isolation       string                           `mapstructure:"isolation" yaml:"isolation,omitempty"`

	spec.Name = svc.Name
	spec.TaskTemplate.ContainerSpec.Args = svc.Command
	spec.TaskTemplate.ContainerSpec.DNSConfig = &swarm.DNSConfig{
		Nameservers: svc.DNS,
		Search:      svc.DNSSearch,
	}
	spec.TaskTemplate.ContainerSpec.Command = svc.Entrypoint
	spec.TaskTemplate.ContainerSpec.Configs = composeConfigsToSwarm(svc)
	spec.TaskTemplate.ContainerSpec.Hosts = composeExtraHostsToSwarm(svc)
	spec.TaskTemplate.ContainerSpec.Hostname = svc.Hostname
	spec.TaskTemplate.ContainerSpec.Image = svc.Image
	spec.TaskTemplate.ContainerSpec.Init = svc.Init
	spec.TaskTemplate.ContainerSpec.Labels = svc.Labels
	spec.EndpointSpec.Ports = composePortsToSwarm(svc)
	spec.TaskTemplate.ContainerSpec.ReadOnly = svc.ReadOnly
	spec.TaskTemplate.ContainerSpec.Secrets = composeSecretsToSwarm(svc)
	spec.TaskTemplate.ContainerSpec.OpenStdin = svc.StdinOpen
	spec.TaskTemplate.ContainerSpec.StopGracePeriod = svc.StopGracePeriod
	spec.TaskTemplate.ContainerSpec.StopSignal = svc.StopSignal
	spec.TaskTemplate.ContainerSpec.TTY = svc.Tty
	spec.TaskTemplate.ContainerSpec.User = svc.User
	spec.TaskTemplate.ContainerSpec.Mounts = composeVolumesToSwarm(svc)
	spec.TaskTemplate.ContainerSpec.Dir = svc.WorkingDir
	spec.TaskTemplate.ContainerSpec.Isolation = container.Isolation(svc.Isolation)
}

func composeConfigsToSwarm(svc *types.ServiceConfig) []*swarm.ConfigReference {
	var configs []*swarm.ConfigReference

	for _, cfg := range svc.Configs {
		var mode os.FileMode
		if cfg.Mode != nil {
			mode = os.FileMode(*cfg.Mode)
		}

		configs = append(configs, &swarm.ConfigReference{
			File: &swarm.ConfigReferenceFileTarget{
				Name: cfg.Target,
				Mode: mode,
				UID:  cfg.UID,
				GID:  cfg.GID,
			},
			ConfigName: cfg.Source,
		})
	}

	return configs
}

func composeExtraHostsToSwarm(svc *types.ServiceConfig) []string {
	var hosts []string

	for _, item := range svc.ExtraHosts {
		// compose: <host>:<ip>
		if v := strings.SplitN(item, ":", 2); len(v) == 2 {
			// swarm: IP-address hostname(s)
			hosts = append(hosts, fmt.Sprintf("%s %s", v[1], v[0]))
		}
	}

	return hosts
}

func composePortsToSwarm(svc *types.ServiceConfig) []swarm.PortConfig {
	var ports []swarm.PortConfig

	for _, p := range svc.Ports {
		ports = append(ports, swarm.PortConfig{
			PublishMode:   swarm.PortConfigPublishMode(p.Mode),
			TargetPort:    p.Target,
			PublishedPort: p.Published,
			Protocol:      swarm.PortConfigProtocol(p.Protocol),
		})
	}

	return ports
}

func composeSecretsToSwarm(svc *types.ServiceConfig) []*swarm.SecretReference {
	var secrets []*swarm.SecretReference

	for _, secret := range svc.Secrets {
		var mode os.FileMode
		if secret.Mode != nil {
			mode = os.FileMode(*secret.Mode)
		}

		secrets = append(secrets, &swarm.SecretReference{
			File: &swarm.SecretReferenceFileTarget{
				Name: secret.Target,
				Mode: mode,
				UID:  secret.UID,
				GID:  secret.GID,
			},
			SecretName: secret.Source,
		})
	}

	return secrets
}

func composeVolumesToSwarm(svc *types.ServiceConfig) []mount.Mount {
	var mounts []mount.Mount

	for _, vol := range svc.Volumes {
		var (
			bindOptions   *mount.BindOptions
			volumeOptions *mount.VolumeOptions
			tmpfsOptions  *mount.TmpfsOptions
		)

		if vol.Bind != nil {
			bindOptions = &mount.BindOptions{
				Propagation: mount.Propagation(vol.Bind.Propagation),
			}
		}

		if vol.Volume != nil {
			volumeOptions = &mount.VolumeOptions{
				NoCopy: vol.Volume.NoCopy,
			}
		}

		if vol.Tmpfs != nil {
			tmpfsOptions = &mount.TmpfsOptions{
				SizeBytes: vol.Tmpfs.Size,
			}
		}

		mounts = append(mounts, mount.Mount{
			Type:        mount.Type(vol.Type),
			Source:      vol.Source,
			Target:      vol.Target,
			ReadOnly:    vol.ReadOnly,
			Consistency: mount.Consistency(vol.Consistency),

			BindOptions:   bindOptions,
			VolumeOptions: volumeOptions,
			TmpfsOptions:  tmpfsOptions,
		})
	}

	return mounts
}
