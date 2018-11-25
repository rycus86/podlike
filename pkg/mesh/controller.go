package mesh

import (
	"fmt"
	"github.com/rycus86/docker-filter/pkg/connect"
	"net"
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
