package healthcheck

import (
	"fmt"
	"net"
)

func Check() bool {
	conn, err := net.Dial(networkType, networkAddress)
	if err != nil {
		fmt.Println("Healthcheck error:", err)
		return false
	}
	defer conn.Close()

	buf := make([]byte, 64)

	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Healthcheck error:", err)
		return false
	}

	value := string(buf[0:n])

	fmt.Println("Healthcheck result:", value)

	return value == getStateName(StateHealthy)
}
