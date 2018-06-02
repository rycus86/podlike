package component

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"strings"
)

func asDeviceMappings(devices []string) ([]container.DeviceMapping, error) {
	if devices == nil {
		return nil, nil
	}

	mapped := make([]container.DeviceMapping, len(devices), len(devices))

	for idx, device := range devices {
		var source, destination, permissions string

		parts := strings.Split(device, ":")
		switch len(parts) {
		case 3:
			permissions = parts[2]
			fallthrough
		case 2:
			destination = parts[1]
			fallthrough
		case 1:
			source = parts[0]
		default:
			return nil, errors.New(fmt.Sprintf("unexpected device mapping: %s", device))
		}

		if destination == "" {
			destination = source
		}

		mapped[idx] = container.DeviceMapping{
			PathOnHost:        source,
			PathInContainer:   destination,
			CgroupPermissions: permissions,
		}
	}

	return mapped, nil
}
