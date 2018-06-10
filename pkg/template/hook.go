package template

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
)

func podTemplateHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f.Kind() == reflect.Slice && t == reflect.TypeOf(podTemplate{}) {

			item := reflect.ValueOf(data).Index(0).Interface()
			return podTemplate{Template: item.(string)}, nil

		} else if f.Kind() == reflect.String {

			if t == reflect.TypeOf(podTemplate{}) {
				return podTemplate{Template: data.(string)}, nil

			} else if t.Kind() == reflect.Slice && t.Elem() == reflect.TypeOf(podTemplate{}) {
				return []podTemplate{
					{Template: data.(string)},
				}, nil
			}

		} else if f.Kind() == reflect.Map {

			item := reflect.ValueOf(data).Interface()
			if m, ok := item.(map[string]interface{}); ok {
				if inline, ok := m[TypeInline]; ok {

					if t == reflect.TypeOf(podTemplate{}) {
						return podTemplate{Template: inline.(string), Inline: true}, nil

					} else if t.Kind() == reflect.Slice && t.Elem() == reflect.TypeOf(podTemplate{}) {
						return []podTemplate{
							{Template: inline.(string), Inline: true},
						}, nil
					}

				} else if httpSource, ok := m[TypeHttp]; ok {

					if t == reflect.TypeOf(podTemplate{}) {
						return podTemplate{Template: httpSource.(string), Http: true}, nil

					} else if t.Kind() == reflect.Slice && t.Elem() == reflect.TypeOf(podTemplate{}) {
						return []podTemplate{
							{Template: httpSource.(string), Http: true},
						}, nil
					}

				}
			}

		}

		return data, nil
	}
}
