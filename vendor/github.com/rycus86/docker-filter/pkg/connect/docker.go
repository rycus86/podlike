package connect

import "net"

func NewProxyForDockerCli(remote func() (net.Conn, error)) *Proxy {
	// TODO the container wait response blocks the attach - it seems
	return NewProxy(remote, "/wait")
}
