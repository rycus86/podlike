package connect

import "net"

func (cp *connectionPair) closeReading() {
	closeRead(cp.localConn.Conn)
	closeWrite(cp.remoteConn)
}

func closeRead(c net.Conn) error {
	if tcpConn, ok := c.(*net.TCPConn); ok {
		return tcpConn.CloseRead()
	} else if unixConn, ok := c.(*net.UnixConn); ok {
		return unixConn.CloseRead()
	} else {
		return nil
	}
}

func closeWrite(c net.Conn) error {
	if tcpConn, ok := c.(*net.TCPConn); ok {
		return tcpConn.CloseWrite()
	} else if unixConn, ok := c.(*net.UnixConn); ok {
		return unixConn.CloseWrite()
	} else {
		return nil
	}
}
