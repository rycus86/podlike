package mesh

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/docker-filter/pkg/connect"
	"net/http"
	"runtime/debug"
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

			processService(req, templateFiles...)

			return req
		})
}
