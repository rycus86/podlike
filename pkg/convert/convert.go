package convert

import (
	"errors"
	"fmt"
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mattn/go-shellwords"
	"strings"
)

func ToStrSlice(value interface{}) (strslice.StrSlice, error) {
	if value == nil {
		return nil, nil
	}

	stringValue, ok := value.(string)
	if ok {
		return shellwords.Parse(stringValue)
	}

	sliceValue, ok := value.([]string)
	if ok {
		return sliceValue, nil
	}

	slice, ok := value.([]interface{})
	if ok {
		values := make([]string, len(slice), len(slice))

		for idx, item := range slice {
			values[idx] = fmt.Sprintf("%s", item)
		}

		return values, nil
	} else {
		return nil, errors.New(fmt.Sprintf("invalid string or slice: %T %+v", value, value))
	}
}

func ToStringSlice(value interface{}) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	if rawSlice, ok := value.([]interface{}); ok {
		slice := make([]string, 0, len(rawSlice))

		for _, item := range rawSlice {
			if asString, ok := item.(string); ok {
				slice = append(slice, asString)
			} else {
				return nil, errors.New(fmt.Sprintf("not a string item: %+v (%T)", item, item))
			}
		}

		return slice, nil
	}

	if asMap, ok := value.(map[interface{}]interface{}); ok {
		slice := make([]string, 0, len(asMap))

		for key, value := range asMap {
			if keyString, ok := key.(string); ok {
				if valueString, ok := value.(string); ok {
					slice = append(slice, keyString+"="+valueString)
				} else {
					return nil, errors.New(fmt.Sprintf("not a string value: %+v (%T)", value, value))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("not a string key: %+v (%T)", key, key))
			}
		}

		return slice, nil
	}

	return nil, errors.New(fmt.Sprintf("unexpected string slice: %+v", value))
}

func ToStringToStringMap(value interface{}) (map[string]string, error) {
	if value == nil {
		return nil, nil
	}

	if asRawMap, ok := value.(map[interface{}]interface{}); ok {
		asMap := map[string]string{}

		for key, value := range asRawMap {
			if keyString, ok := key.(string); ok {
				if valueString, ok := value.(string); ok {
					asMap[keyString] = valueString
				} else {
					return nil, errors.New(fmt.Sprintf("not a string value: %+v (%T)", value, value))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("not a string key: %+v (%T)", key, key))
			}
		}

		return asMap, nil
	}

	if slice, ok := value.([]interface{}); ok {
		asMap := map[string]string{}

		for _, rawItem := range slice {
			if item, ok := rawItem.(string); ok {
				parts := strings.SplitN(item, "=", 2)

				if len(parts) == 2 {
					asMap[parts[0]] = parts[1]
				} else {
					asMap[item] = ""
				}
			} else {
				return nil, errors.New(fmt.Sprintf("not a string item: %+v (%T)", item, item))
			}
		}

		return asMap, nil
	}

	return nil, errors.New(fmt.Sprintf("unexpected string slice: %+v", value))
}

func SwarmSpecToComposeService(spec *swarm.ServiceSpec) types.ServiceConfig {
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

	// TODO
	return types.ServiceConfig{
		Name: spec.Name,
		//Name string `yaml:"-"`

		//Build           BuildConfig                      `yaml:",omitempty"`
		//CapAdd          []string                         `mapstructure:"cap_add" yaml:"cap_add,omitempty"`
		//CapDrop         []string                         `mapstructure:"cap_drop" yaml:"cap_drop,omitempty"`
		//CgroupParent    string                           `mapstructure:"cgroup_parent" yaml:"cgroup_parent,omitempty"`
		Command: types.ShellCommand(spec.TaskTemplate.ContainerSpec.Args),
		//Command         ShellCommand                     `yaml:",omitempty"`

		//Configs         []ServiceConfigObjConfig         `yaml:",omitempty"`
		//ContainerName   string                           `mapstructure:"container_name" yaml:"container_name,omitempty"`
		//CredentialSpec  CredentialSpecConfig             `mapstructure:"credential_spec" yaml:"credential_spec,omitempty"`
		//DependsOn       []string                         `mapstructure:"depends_on" yaml:"depends_on,omitempty"`
		//Deploy          DeployConfig                     `yaml:",omitempty"`
		//Devices         []string                         `yaml:",omitempty"`
		//DNS             StringList                       `yaml:",omitempty"`
		//DNSSearch       StringList                       `mapstructure:"dns_search" yaml:"dns_search,omitempty"`
		//DomainName      string                           `mapstructure:"domainname" yaml:"domainname,omitempty"`
		Entrypoint: types.ShellCommand(spec.TaskTemplate.ContainerSpec.Command),
		//Entrypoint      ShellCommand                     `yaml:",omitempty"`

		//Environment     MappingWithEquals                `yaml:",omitempty"`
		//EnvFile         StringList                       `mapstructure:"env_file" yaml:"env_file,omitempty"`
		//Expose          StringOrNumberList               `yaml:",omitempty"`
		//ExternalLinks   []string                         `mapstructure:"external_links" yaml:"external_links,omitempty"`
		//ExtraHosts      HostsList                        `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty"`
		//Hostname        string                           `yaml:",omitempty"`
		//HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
		//Image           string                           `yaml:",omitempty"`
		Image: spec.TaskTemplate.ContainerSpec.Image,

		//Ipc             string                           `yaml:",omitempty"`
		//Labels          Labels                           `yaml:",omitempty"`
		Labels: types.Labels(spec.TaskTemplate.ContainerSpec.Labels),

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
		Tty: spec.TaskTemplate.ContainerSpec.TTY,

		//Ulimits         map[string]*UlimitsConfig        `yaml:",omitempty"`
		//User            string                           `yaml:",omitempty"`
		//Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
		Volumes: volumes,

		//WorkingDir      string                           `mapstructure:"working_dir" yaml:"working_dir,omitempty"`
		//Isolation       string                           `mapstructure:"isolation" yaml:"isolation,omitempty"`
	}
}

func MergeComposeServiceIntoSwarmSpec(svc types.ServiceConfig, spec *swarm.ServiceSpec) {
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
