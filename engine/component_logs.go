package engine

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"strings"
)

func (c *Component) streamLogs() {
	if reader, err := c.client.api.ContainerLogs(context.Background(), c.containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}); err == nil {
		defer reader.Close()

		fmt.Println("Streaming logs for", c.Name)

		bufReader := bufio.NewReader(reader)
		defer reader.Close()

		for {
			out, _, err := bufReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					fmt.Println("Stopped streaming logs for", c.Name, ":", err)
				}
				return
			}

			streamType := "out"
			if out[0] == 2 {
				streamType = "err"
			}

			fmt.Printf("[%s] %s: %s\n", streamType, c.Name, strings.TrimSpace(string(out[8:])))
		}
	}
}
