package template

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"reflect"
	"regexp"
)

var (
	httpPrefix = regexp.MustCompile("(?i)^https?://")
)

func podTemplateHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f.Kind() == reflect.Slice {
			return hookFromSlice(t, data), nil

		} else if f.Kind() == reflect.String {
			return hookFromString(t, data), nil

		} else if f.Kind() == reflect.Map {
			return hookFromMap(t, data)

		} else {
			return data, nil

		}
	}
}

func hookFromSlice(t reflect.Type, data interface{}) interface{} {
	if t == reflect.TypeOf(podTemplate{}) {
		item := reflect.ValueOf(data).Index(0).Interface()
		return hookStringToPodTemplate(item.(string))
	}

	return data
}

func hookFromString(t reflect.Type, data interface{}) interface{} {
	if t == reflect.TypeOf(podTemplate{}) {
		return hookStringToPodTemplate(data.(string))

	} else if t.Kind() == reflect.Slice && t.Elem() == reflect.TypeOf(podTemplate{}) {
		return []podTemplate{
			hookStringToPodTemplate(data.(string)),
		}
	}

	return data
}

func hookStringToPodTemplate(source string) podTemplate {
	if httpPrefix.MatchString(source) {
		return podTemplate{Http: &httpTemplate{URL: source}}
	} else {
		return podTemplate{File: &fileTemplate{Path: source}}
	}
}

func hookFromMap(t reflect.Type, data interface{}) (interface{}, error) {
	item := reflect.ValueOf(data).Interface()
	if m, ok := item.(map[string]interface{}); ok {
		if inline, ok := m[TypeInline]; ok {
			return hookFromInlineConfig(t, inline)

		} else if httpSource, ok := m[TypeHttp]; ok {
			return hookFromHttpConfig(t, httpSource)

		}
	}

	return data, nil
}

func hookFromInlineConfig(t reflect.Type, inline interface{}) (interface{}, error) {
	// convert an inline mapping into a YAML string
	if inlineMap, ok := inline.(map[string]interface{}); ok {
		if content, err := yaml.Marshal(inlineMap); err != nil {
			return nil, err
		} else {
			inline = string(content)
		}
	}

	if t == reflect.TypeOf(podTemplate{}) {
		return podTemplate{Inline: inline.(string)}, nil

	} else if t.Kind() == reflect.Slice && t.Elem() == reflect.TypeOf(podTemplate{}) {
		return []podTemplate{
			{Inline: inline.(string)},
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("invalid `inline` config type: %T (%+v)", inline, inline))
}

func hookFromHttpConfig(t reflect.Type, source interface{}) (interface{}, error) {
	var (
		url      string
		insecure bool
		fallback *podTemplate
	)

	if src, ok := source.(string); ok {
		url = src

	} else if config, ok := source.(map[string]interface{}); ok {
		if v, ok := config[PropURL]; !ok {
			return nil, errors.New(fmt.Sprintf("missing `url` property on %+v", config))
		} else if src, ok := v.(string); !ok {
			return nil, errors.New(fmt.Sprintf("invalid `url` type %T on %+v", v, config))
		} else {
			url = src
		}

		if v, ok := config[PropInsecure]; ok {
			if i, ok := v.(bool); !ok {
				return nil, errors.New(fmt.Sprintf("invalid `insecure` type %T on %+v", v, config))
			} else {
				insecure = i
			}
		}

		if v, ok := config[PropFallback]; ok {
			var f podTemplate

			if decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				Result:     &f,
				DecodeHook: podTemplateHookFunc(),
			}); err != nil {
				return nil, err
			} else if err := decoder.Decode(v); err != nil {
				return nil, err
			} else {
				fallback = &f
			}
		}
	}

	if t == reflect.TypeOf(podTemplate{}) {
		return podTemplate{
			Http: &httpTemplate{
				URL:      url,
				Insecure: insecure,
				Fallback: fallback,
			},
		}, nil

	} else if t.Kind() == reflect.Slice && t.Elem() == reflect.TypeOf(podTemplate{}) {
		return []podTemplate{
			{Http: &httpTemplate{URL: url, Insecure: insecure, Fallback: fallback}},
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("invalid `http` config type: %T (%+v)", source, source))
}
