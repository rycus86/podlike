package engine

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/mattn/go-shellwords"
	"strings"
)

func asStrSlice(value interface{}) (strslice.StrSlice, error) {
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

func asStringSlice(value interface{}) ([]string, error) {
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

func asStringToStringMap(value interface{}) (map[string]string, error) {
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
