package convert

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/strslice"
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
