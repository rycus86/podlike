package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/schema"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/versions"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

func (ts *transformSession) prepareConfiguration(configFiles []types.ConfigFile) {
	for _, configFile := range configFiles {
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
		panic(fmt.Sprintf("top level services is not a mapping, but a %T\n%+v", services, services))
	}

	for serviceName, definition := range mServices {
		if definition == nil {
			continue
		}

		mDefinition, ok := definition.(map[string]interface{})
		if !ok {
			panic(fmt.Sprintf(
				"service definition is not a mapping for %s, but a %T\n%+v",
				serviceName, definition, definition))
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
			panic(fmt.Sprintf("failed to set up a podlike config decoder : %s", err.Error()))
		}

		err = decoder.Decode(configSection)
		if err != nil {
			panic(fmt.Sprintf("failed to decode a podlike configuration : %s\n%+v", err.Error(), configSection))
		}

		ts.registerService(serviceName, config)
	}
}

func (ts *transformSession) collectTopLevelConfigurations(configFile types.ConfigFile) {
	if configSection, ok := configFile.Config[XPodlikeExtension]; ok {
		globalConfig, ok := configSection.(map[string]interface{})
		if !ok {
			panic(fmt.Sprintf(
				"top level x-podlike config is not a mapping, but a %T\n%+v",
				configSection, configSection))
		}

		// extract the top level global arguments first
		if args, ok := globalConfig[ArgsProperty]; ok {
			if mArgs, ok := args.(map[string]interface{}); ok {
				mergeRecursively(ts.Args, mArgs)
				delete(globalConfig, ArgsProperty)
			} else if args != nil {
				panic(fmt.Sprintf("template args is not a mapping, but a %T\n%+v", args, args))
			}
		}

		// parse the rest of the configuration as {service -> config} map
		var configs map[string]transformConfiguration

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:     &configs,
			DecodeHook: podTemplateHookFunc(),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to set up a podlike config decoder : %s", err.Error()))
		}

		err = decoder.Decode(configSection)
		if err != nil {
			panic(fmt.Sprintf(
				"failed to decode the top-level podlike configurations : %s\n%+v",
				err.Error(), configSection))
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
		panic(fmt.Sprintf(
			"failed to convert a template transformer session to YAML : %s\n%+v",
			err.Error(), ts.Project))
	}

	return string(output)
}
