package mesh

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/docker-filter/pkg/connect"
	"github.com/rycus86/podlike/pkg/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

func setupFilters(proxy *connect.Proxy, templateFiles ...string) {
	proxy.FilterResponses("/services/[^/]+$", filterServiceInspect())

	proxy.FilterRequests("/services/create", filterServiceCreate(templateFiles...))
	// TODO do these two need their own method?
	proxy.FilterRequests("/services/update", filterServiceCreate(templateFiles...))
	proxy.FilterRequests("/services/.+/update", filterServiceCreate(templateFiles...))
}

func filterServiceInspect() connect.ResponseFilterFunc {
	inspectFilter := connect.FilterResponseAsJson(func() connect.T { return &swarm.Service{} },
		func(r connect.T) connect.T {
			defer func() {
				if e := recover(); e != nil {
					fmt.Println("oops:", e) // TODO
					fmt.Println(string(debug.Stack()))
				}
			}()

			resp := r.(*swarm.Service)

			if originalTemplate := resp.Spec.Labels["podlike.mesh.original.template"]; originalTemplate != "" {
				var original swarm.TaskSpec
				if err := json.Unmarshal([]byte(originalTemplate), &original); err != nil {
					panic(err)
				}

				resp.Spec.TaskTemplate = original
			}
			if originalEndpointSpec := resp.Spec.Labels["podlike.mesh.original.endpoint-spec"]; originalEndpointSpec != "" {
				var original swarm.EndpointSpec
				if err := json.Unmarshal([]byte(originalEndpointSpec), &original); err != nil {
					panic(err)
				}

				resp.Spec.EndpointSpec = &original
			} else {
				resp.Spec.EndpointSpec = nil
			}

			return resp
		})

	return func(resp *http.Response, body []byte) (*http.Response, error) {
		if resp.Request == nil {
			return nil, nil
		}

		if resp.Request.Method != "GET" {
			return nil, nil
		}

		return inspectFilter(resp, body)
	}
}

func filterServiceCreate(templateFiles ...string) connect.RequestFilterFunc {
	return connect.FilterRequestAsJson(func() connect.T { return &swarm.ServiceSpec{} },
		func(r connect.T) connect.T {
			defer func() {
				if e := recover(); e != nil {
					fmt.Println("oops:", e) // TODO
					fmt.Println(string(debug.Stack()))
				}
			}()

			req := r.(*swarm.ServiceSpec)

			// TODO document why these are needed
			if originalTemplate, err := json.Marshal(req.TaskTemplate); err != nil {
				panic(err)
			} else {
				req.Labels["podlike.mesh.original.template"] = string(originalTemplate)
			}
			if req.EndpointSpec != nil {
				if originalEndpointSpec, err := json.Marshal(req.EndpointSpec); err != nil {
					panic(err)
				} else {
					req.Labels["podlike.mesh.original.endpoint-spec"] = string(originalEndpointSpec)
				}
			}

			name := req.Name
			req.Name = "app"
			svc := convertSwarmSpecToComposeService(req)

			// TODO this might need caching to download remote templates
			// TODO also timeouts, retries, etc.
			tmplFiles, tempFiles, err := ensureTemplateFiles(templateFiles)
			defer func() {
				for _, tf := range tempFiles {
					os.Remove(tf)
				}
			}()

			if err != nil {
				panic(err)
			}

			ts := template.NewSession(tmplFiles...)
			ts.ReplaceService(&svc)
			ts.Execute()
			ts.Project.Services[0].Name = name

			mergeComposeServiceIntoSwarmSpec(&ts.Project.Services[0], req)

			return req
		})
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
