package component

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func (c *Component) streamLogs() {
	if reader, err := c.engine.StreamLogs(c.container.ID); err == nil {
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

			if len(out) < 8 {
				continue
			}

			streamType := "out"
			if out[0] == 2 {
				streamType = "err"
			}

			fmt.Printf("[%s] %s: %s\n", streamType, c.Name, strings.TrimSpace(string(out[8:])))
		}
	}
}
