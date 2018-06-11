package template

import (
	"fmt"
	"github.com/rycus86/podlike/pkg/convert"
)

func mergeServiceProperties(target, source map[string]interface{}, mergeKeys []string) {
	serviceConfig := asSingleServiceConfig(target)

	for key, value := range source {
		if !containsItem(mergeKeys, key) {
			continue
		}

		if _, alreadyExists := serviceConfig[key]; !alreadyExists {
			serviceConfig[key] = value
		}
	}
}

func asSingleServiceConfig(services map[string]interface{}) map[string]interface{} {
	for _, svc := range services {
		return svc.(map[string]interface{})
	}

	panic(fmt.Sprintf("no service configuration found in %+v", services))
}

func containsItem(coll []string, item string) bool {
	for _, i := range coll {
		if i == item {
			return true
		}
	}

	return false
}

// Merges to YAML configurations, represented as `map[string]interface{}`.
// If needed, slices are converted to maps automatically (labels for example),
// by treating their items as a key-value string with an equals sign in the middle.
// If we're trying to merge a map into a slice, the slice is automatically
// converted to a map before merging, following the same rules for key-value items.
// We can also merge a string into a slice, but not the other way around.
// Existing items won't get overwritten, apart from the custom slice/map merge logic just described.
func mergeRecursively(target, source map[string]interface{}) {
	for key, value := range source {
		if existing, ok := target[key]; ok {
			if m, ok := existing.(map[string]interface{}); ok {

				if v, ok := value.(map[string]interface{}); ok {
					// merge map into map
					mergeRecursively(m, v)

				} else if v, ok := value.([]interface{}); ok {
					// merge slice into map
					siMap := map[string]interface{}{}
					if ssMap, err := convert.ToStringToStringMap(v); err == nil {
						for sKey, sValue := range ssMap {
							siMap[sKey] = sValue
						}
					}
					mergeRecursively(m, siMap)
				}

			} else if s, ok := existing.([]interface{}); ok {

				if v, ok := value.([]interface{}); ok {
					// merge slice into slice
					for _, item := range v {
						s = append(s, item)
					}

				} else if v, ok := value.(map[string]interface{}); ok {
					// merge map into slice by converting them to a map
					siMap := map[string]interface{}{}
					if ssMap, err := convert.ToStringToStringMap(s); err == nil {
						for sk, sv := range ssMap {
							siMap[sk] = sv
						}
						mergeRecursively(siMap, v)
						target[key] = siMap
					}

					// if the above failed, skip this item
					continue
				} else if v, ok := value.(string); ok {
					// merge string into slice
					s = append(s, v)
				}

				target[key] = s
			}
		} else {
			target[key] = value
		}
	}
}

var (
	mergedPodKeys = []string{
		"configs",
		"credential_spec",
		"deploy",
		"dns",
		"dns_search",
		"domainname",
		"entrypoint",
		"env_file",
		"environment",
		"expose",
		"extra_hosts",
		"healthcheck",
		"hostname",
		"ipc",
		"isolation",
		"labels",
		"logging",
		"mac_address",
		"networks",
		"pid",
		"ports",
		"privileged",
		"read_only",
		"secrets",
		"shm_size",
		"stdin_open",
		"stop_grace_period",
		"tmpfs",
		"tty",
		"ulimits",
		"user",
		"volumes",
		"working_dir",
	}

	mergedTransformerKeys = []string{
		"blkio_config",
		"cap_add",
		"cap_drop",
		"command",
		"cpu_count",
		"cpu_percent",
		"cpu_period",
		"cpu_quota",
		"cpu_rt_period",
		"cpu_rt_runtime",
		"cpu_shares",
		"cpus",
		"cpuset",
		"depends_on",
		"device_cgroup_rules",
		"devices",
		"entrypoint",
		"env_file",
		"environment",
		"group_add",
		"healthcheck",
		"image",
		"isolation",
		"labels",
		"logging",
		"mem_limit",
		"mem_reservation",
		"mem_swappiness",
		"memswap_limit",
		"oom_kill_disable",
		"oom_score_adj",
		"pids_limit",
		"privileged",
		"read_only",
		"runtime",
		"security_opt",
		"shm_size",
		"stdin_open",
		"stop_grace_period",
		"stop_signal",
		"storage_opt",
		"sysctls",
		"tmpfs",
		"tty",
		"ulimits",
		"user",
		"userns_mode",
		"volumes",
		"working_dir",
	}
)
