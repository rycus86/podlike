package mesh

import (
	"fmt"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/docker-filter/pkg/connect"
	"github.com/rycus86/podlike/pkg/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

func StartMeshController(args ...string) {
	config := configure(args...)
	runController(config)
}

func runController(config *Configuration) {
	var listeners []net.Listener

	defer func() {
		for _, listener := range listeners {
			listener.Close()
		}
	}()

	for _, listenAddress := range config.ListenAddresses {
		network, address := parseNetworkAndAddress(listenAddress)

		if listener, err := net.Listen(network, address); err != nil {
			panic(fmt.Errorf("failed to start listener: %s", err))
		} else {
			listeners = append(listeners, listener)
		}
	}

	engineNetwork, engineAddress := parseNetworkAndAddress(config.EngineConnection)

	proxy := connect.NewProxyForDockerCli(func() (net.Conn, error) {
		return net.Dial(engineNetwork, engineAddress)
	})

	for idx, listener := range listeners {
		proxy.AddListener(fmt.Sprintf("L%02d", idx+1), listener) // TODO prefix
	}

	setupFilters(proxy, config.Templates...)

	panic(proxy.Process())
}

func processService(spec *swarm.ServiceSpec, templateFiles ...string) {
	var tempFilesToRemove []string
	defer func() {
		for _, tf := range tempFilesToRemove {
			os.Remove(tf)
		}
	}()

	// TODO this might need caching to download remote templates
	// TODO also timeouts, retries, etc.
	tmplFiles, tempFiles, err := ensureTemplateFiles(templateFiles)
	tempFilesToRemove = append(tempFilesToRemove, tempFiles...)

	if err != nil {
		panic(err)
	}

	ts := template.NewSession(tmplFiles...) // TODO maybe this needs to be told not to add default pod/transformers just yet
	tc := template.TC{
		Args: map[string]interface{}{},
	}

	for _, templateName := range strings.Split(spec.Labels["podlike.mesh.templates"], ",") {
		tName := strings.TrimSpace(templateName)
		if len(tName) == 0 {
			tName = "default"
		}

		if cfg, ok := ts.Configurations[tName]; !ok {
			continue // skip non-existing configuration
		} else {
			for _, pod := range cfg.Pod {
				if pod.IsDefault { // TODO
					continue
				}

				tc.Pod = append(tc.Pod, pod)
			}

			for _, transformer := range cfg.Transformer {
				if transformer.IsDefault { // TODO
					continue
				}

				tc.Transformer = append(tc.Transformer, transformer)
			}

			tc.Init = append(tc.Init, cfg.Init...)
			tc.Templates = append(tc.Templates, cfg.Templates...)
			tc.Copy = append(tc.Copy, cfg.Copy...)

			for key, value := range cfg.Args {
				tc.Args[key] = value // TODO merge rules?
			}
		}
	}

	name := spec.Name

	spec.Name = "$svc"
	svc := convertSwarmSpecToComposeService(spec)

	tc.Service = &svc
	ts.TODO_Set(spec.Name, tc) // TODO
	ts.Project.Services = append(ts.Project.Services, svc)

	ts.Execute()

	changed := findServiceByName(ts.Project.Services, spec.Name)
	changed.Name = name

	mergeComposeServiceIntoSwarmSpec(changed, spec)

	spec.Name = name
}

func findServiceByName(services types.Services, name string) *types.ServiceConfig {
	for _, svc := range services {
		if svc.Name == name {
			return &svc
		}
	}

	return nil
}

func ensureTemplateFiles(templateFiles []string) ([]string, []string, error) {
	var templates []string
	var tempFiles []string
	var templateError error

	for _, tmpl := range templateFiles {
		if fi, err := os.Stat(tmpl); err == nil && !fi.IsDir() {
			templates = append(templates, tmpl)
			continue
		}

		// TODO maybe the template engine should support URLs directly
		if strings.HasPrefix(strings.ToLower(tmpl), "http://") ||
			strings.HasPrefix(strings.ToLower(tmpl), "https://") {

			if resp, err := http.Get(tmpl); err != nil {
				templateError = fmt.Errorf("failed to fetch a template from %s: %s", tmpl, err)
			} else if resp.StatusCode != 200 {
				resp.Body.Close()
				templateError = fmt.Errorf("failed to fetch a template from %s: HTTP %d", tmpl, resp.StatusCode)
			} else if f, err := ioutil.TempFile("", "podlike.mesh.*.yml"); err != nil {
				resp.Body.Close()

				templateError = fmt.Errorf("failed to create a temporary file for %s: %s", tmpl, err)
			} else if _, err := io.Copy(f, resp.Body); err != nil {
				os.Remove(f.Name())
				resp.Body.Close()

				templateError = fmt.Errorf("failed to write temporary file at %s for %s: %s", f.Name(), tmpl, err)
			} else {
				resp.Body.Close()

				templates = append(templates, f.Name())
				tempFiles = append(tempFiles, f.Name())
			}

		} else {
			templateError = fmt.Errorf("template not found at %s", tmpl)

		}

		if templateError != nil {
			break
		}
	}

	return templates, tempFiles, templateError
}
