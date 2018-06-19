package template

import (
	"fmt"
	"github.com/docker/cli/cli/compose/types"
)

// Takes the paths of one or more Compose files,
// loads and transforms them using the templates,
// and returns the resulting YAML as string.
func Transform(inputFiles ...string) string {
	session := newSession(inputFiles...)

	for serviceName, config := range session.Configurations {
		index := config.getServiceIndex()
		if index < 0 {
			panic(fmt.Sprintf("service index not found for %s\n%+v", serviceName, session.Project.Services))
		}

		podController := executePodTemplates(&config)

		mcName, mainComponent := executeTransformers(&config)
		podController.Labels["pod.component."+mcName] = mainComponent

		for cName, component := range executeTemplates(&config) {
			podController.Labels["pod.component."+cName] = component
		}

		for cName, cp := range executeCopyTemplates(&config) {
			podController.Labels["pod.copy."+cName] = cp
		}

		session.Project.Services[index] = podController
	}

	return session.toYamlString()
}

// Renders the main pod template for the controller, then merges the
// results of any add-ons (templates other than the first one) into it.
// The root key in the final result is always going to be the name of the service,
// regardless of what root keys the templates define.
// The add-ons won't overwrite already existing properties, although they can
// extend them, see `mergeRecursively` in `merge.go`.
// The existing properties of the original service definition are also copied
// over for most keys, see `mergedPodKeys` in `merge.go`, but only if
// the template didn't create them on its own.
func executePodTemplates(tc *transformConfiguration) types.ServiceConfig {
	definition := map[string]interface{}{}

	for _, tmpl := range tc.Pod {
		rendered := tmpl.render(tc)
		if len(rendered) != 1 {
			panic(fmt.Sprintf(
				"the pod template can only define a single controller service, but got %d\n%+v",
				len(rendered), rendered))
		}

		rendered = changeRootKey(rendered, tc.Service.Name)
		mergeRecursively(definition, rendered)
	}

	mergeServiceProperties(definition, tc.getServiceConfig(), mergedPodKeys)

	// add in some defaults if still missing at this point (image and volumes, for example)
	mergeRecursively(definition, getMinimalPodProperties(tc.Service.Name))

	converted := convertToServices(definition, tc.Session.WorkingDir)
	if len(converted) != 1 {
		panic(fmt.Sprintf(
			"somehow we ended up with %d definitions for the controller\n%+v",
			len(converted), converted))
	}

	pod := converted[0]

	// ensure the result has labels to add the components to
	if pod.Labels == nil {
		pod.Labels = types.Labels{}
	}

	return pod
}

// Similarly to how the controller definition is generated (in `executePodTemplates`),
// the main component is generated using the first template, then any other templates
// are treated as add-ons, where any extra configuration from them is going to be
// merged in, but they can't overwrite existing properties, see `mergeRecursively` in `merge.go`
// The name of the main component is set to the root key of the first template.
// The existing properties of the original service definition are copied
// over for most keys to the component, see `mergedTransformerKeys` in `merge.go`,
// but only if the template didn't create them on its own.
func executeTransformers(tc *transformConfiguration) (string, string) {
	var (
		rootKey    string
		definition = map[string]interface{}{}
	)

	for idx, tmpl := range tc.Transformer {
		rendered := tmpl.render(tc)
		if len(rendered) != 1 {
			panic(fmt.Sprintf(
				"the transformer template can only define a single component, but we got %d\n%+v",
				len(rendered), rendered))
		}

		if idx == 0 {
			rootKey = getRootKey(rendered)
		} else {
			rendered = changeRootKey(rendered, rootKey)
		}

		mergeRecursively(definition, rendered)
	}

	mergeServiceProperties(definition, tc.getServiceConfig(), mergedTransformerKeys)

	// we need to remove depends_on here, the Compose v2 compatible format won't parse for the v3 service type
	servicesWithDependsOn := extractDependsOnConfig(definition)

	converted := convertToServices(definition, tc.Session.WorkingDir)
	if len(converted) != 1 {
		panic(fmt.Sprintf(
			"somehow we ended up with %d definitions for the main component\n%+v",
			len(converted), converted))
	}

	comp := convertToYaml(converted[0])

	// add back the removed depends_on
	comp = insertDependsOnConfig(comp, servicesWithDependsOn, converted[0].Name)

	return converted[0].Name, comp
}

// Renders all the definitions from all the listed templates, and merges them into
// a YAML map with possibly multiple root keys, one for each additional component definition.
// The names of the components come from these root keys, and templates can extend
// the results of the earlier template, if these root keys match.
func executeTemplates(tc *transformConfiguration) map[string]string {
	var (
		definition = map[string]interface{}{}
		components = map[string]string{}
	)

	for _, tmpl := range tc.Templates {
		mergeRecursively(definition, tmpl.render(tc))
	}

	// we need to remove depends_on here, the Compose v2 compatible format won't parse for the v3 service type
	servicesWithDependsOn := extractDependsOnConfig(definition)

	for _, comp := range convertToServices(definition, tc.Session.WorkingDir) {
		rendered := convertToYaml(comp)

		// add back the removed depends_on
		rendered = insertDependsOnConfig(rendered, servicesWithDependsOn, comp.Name)

		components[comp.Name] = rendered
	}

	return components
}

// Renders the copy templates to copy definitions, represented as a slice of
// from/to paths separated by a colon. In the templates, they can be defined
// as a single string, a slice or a map.
func executeCopyTemplates(tc *transformConfiguration) map[string]string {
	var (
		definition = map[string]interface{}{}
		copies     = map[string]string{}
	)

	for _, tmpl := range tc.Copy {
		item := tmpl.render(tc)

		// convert any copy configurations given as maps to slices
		for svc, cps := range item {
			if m, ok := cps.(map[string]interface{}); ok {

				var cp []interface{}

				for k, v := range m {
					cp = append(cp, fmt.Sprintf("%s:%s", k, v))
				}

				item[svc] = cp

			} else if str, ok := cps.(string); ok {

				item[svc] = []interface{}{str}

			}
		}

		mergeRecursively(definition, item)
	}

	for svc, items := range definition {
		copies[svc] = convertToYaml(items)
	}

	return copies
}

func getRootKey(m map[string]interface{}) string {
	for key := range m {
		return key
	}

	panic(fmt.Sprintf("cannot find the root key in an empty map : %+v", m))
}

func changeRootKey(m map[string]interface{}, key string) map[string]interface{} {
	changed := map[string]interface{}{}
	for _, value := range m {
		changed[key] = value
	}
	return changed
}
