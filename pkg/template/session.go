package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/schema"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/versions"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

func (ts *transformSession) prepareConfiguration() {
	for _, configFile := range ts.ConfigFiles {
		ts.collectServiceLevelConfigurations(configFile)
		ts.collectTopLevelConfigurations(configFile)
	}
}

func (ts *transformSession) collectServiceLevelConfigurations(configFile types.ConfigFile) {
	services, ok := configFile.Config[ServicesProperty]
	if !ok {
		return // ok, some YAMLs might only define volumes and such
	}

	mServices, ok := services.(map[string]interface{})
	if !ok {
		panic("top level services is not a mapping")
	}

	for serviceName, definition := range mServices {
		mDefinition, ok := definition.(map[string]interface{})
		if !ok {
			panic(fmt.Sprintf("service definition is not a mapping for %s", serviceName))
		}

		configSection, ok := mDefinition[XPodlikeExtension]
		if !ok {
			continue
		}

		if v := schema.Version(configFile.Config); versions.LessThan(v, "3.7") {
			// we have to delete the extension key below schema version 3.7
			delete(mDefinition, XPodlikeExtension)
		}

		var config transformConfiguration

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:     &config,
			DecodeHook: podTemplateHookFunc(),
		})
		if err != nil {
			panic(err)
		}

		err = decoder.Decode(configSection)
		if err != nil {
			panic(err)
		}

		ts.registerService(serviceName, config)
	}
}

func (ts *transformSession) collectTopLevelConfigurations(configFile types.ConfigFile) {
	if configSection, ok := configFile.Config[XPodlikeExtension]; ok {
		globalConfig, ok := configSection.(map[string]interface{})
		if !ok {
			panic("top level x-podlike config is not a mapping")
		}

		// extract the top level global arguments first
		if args, ok := globalConfig[ArgsProperty]; ok {
			if mArgs, ok := args.(map[string]interface{}); ok {
				mergeRecursively(ts.Args, mArgs)
				delete(globalConfig, ArgsProperty)
			} else {
				panic("template args is not a mapping")
			}
		}

		// parse the rest of the configuration as {service -> config} map
		var configs map[string]transformConfiguration

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:     &configs,
			DecodeHook: podTemplateHookFunc(),
		})
		if err != nil {
			panic(err)
		}

		err = decoder.Decode(configSection)
		if err != nil {
			panic(err)
		}

		for serviceName, config := range configs {
			ts.registerService(serviceName, config)
		}
	}
}

func (ts *transformSession) registerService(name string, config transformConfiguration) {
	if existing, ok := ts.Configurations[name]; ok {
		for _, pod := range config.Pod {
			existing.Pod = append(existing.Pod, pod)
		}

		for _, transformer := range config.Transformer {
			existing.Transformer = append(existing.Transformer, transformer)
		}

		for _, tmpl := range config.Templates {
			existing.Templates = append(existing.Templates, tmpl)
		}

		for _, cp := range config.Copy {
			existing.Copy = append(existing.Copy, cp)
		}

		mergeRecursively(existing.Args, config.Args)

		return
	}

	config.Session = ts

	if len(config.Pod) == 0 {
		config.Pod = []podTemplate{
			getDefaultPodTemplate(),
		}
	}

	if len(config.Transformer) == 0 {
		config.Transformer = []podTemplate{
			getDefaultTransformerTemplate(),
		}
	}

	ts.Configurations[name] = config
}

func (ts *transformSession) toYamlString() string {
	output, err := yaml.Marshal(ts.Project)
	if err != nil {
		panic(err)
	}

	return string(output)
}
