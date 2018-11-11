package mesh

import (
	"github.com/rycus86/docker-filter/pkg/connect"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
)

var (
	mockDockerServer *httptest.Server
	mockListener     net.Listener
	mockProxy        *connect.Proxy
)

func initMocks(handler func(http.ResponseWriter, *http.Request)) error {
	mockDockerServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_ping" {
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.Header().Add("Api-Version", "1.39")
			w.Header().Add("Ostype", "linux")
			w.Header().Add("Server", "Docker/18.09.0 (linux)")
			w.WriteHeader(200)
			w.Write([]byte("OK"))
			return
		}

		handler(w, r)
	}))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	mockListener = listener

	proxy := connect.NewProxy(func() (net.Conn, error) {
		return net.Dial(mockDockerServer.Listener.Addr().Network(), mockDockerServer.Listener.Addr().String())
	})
	proxy.AddListener("test", listener)

	go proxy.Process()

	mockProxy = proxy

	return nil
}

func closeMocks() {
	if mockListener != nil {
		mockListener.Close()
	}
	if mockDockerServer != nil {
		mockDockerServer.Close()
	}
}

func hasDockerCli() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

func runDockerCliCommand(args ...string) {
	cmdArgs := []string{"-H", "tcp://" + mockListener.Addr().String()}

	for _, arg := range args {
		for _, part := range strings.Split(arg, " ") {
			if part == "$jsonFmt" {
				part = "{{ json . }}"
			}

			cmdArgs = append(cmdArgs, part)
		}
	}

	cmd := exec.Command("docker", cmdArgs...)

	//wOut := new(bytes.Buffer)
	//wErr := new(bytes.Buffer)
	//
	//cmd.Stdout = wOut
	//cmd.Stderr = wErr

	cmd.Run()

	//err := cmd.Run()
	//
	//stdout := wOut.String()
	//stderr := wErr.String()
	//
	//fmt.Println("Run:", cmd.Args)
	//fmt.Println("Err:", err)
	//fmt.Println("StdOut:", stdout)
	//fmt.Println("StdErr:", stderr)

	return
}
