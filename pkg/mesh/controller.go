package mesh

import (
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"github.com/rycus86/docker-filter/pkg/connect"
	"github.com/rycus86/podlike/pkg/template"
	"net"
	"runtime/debug"
)

func TestMe() {
	listener, _ := net.Listen("tcp", ":8888")
	defer listener.Close()

	p := connect.NewProxy(func() (conn net.Conn, e error) {
		return net.Dial("unix", "/var/run/docker.sock")
	})
	p.AddListener("", listener)

	p.Handle("/services/create", processServiceCreateRequests("cmd/mesh/for-mesh.yml"))

	panic(p.Process())
}

func processServiceCreateRequests(templateFile string) connect.FilterFunc {
	return connect.FilterAsJson(func() connect.T { return &swarm.ServiceSpec{} },
		func(r connect.T) connect.T {
			defer func() {
				if e := recover(); e != nil {
					fmt.Println("oops:", e)
					fmt.Println(string(debug.Stack()))
				}
			}()

			req := r.(*swarm.ServiceSpec)

			name := req.Name
			req.Name = "app"
			svc := convertSwarmSpecToComposeService(req)

			ts := template.NewSession(templateFile)
			ts.ReplaceService(&svc)
			ts.Execute()
			ts.Project.Services[0].Name = name

			mergeComposeServiceIntoSwarmSpec(ts.Project.Services[0], req)

			return req
		})
}
