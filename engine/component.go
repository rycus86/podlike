package engine

import "fmt"

func (c *Component) init(name string, client *Client) {
	c.Name = name
	c.client = client

	c.warnForSettings()
}

func (c *Component) warnForSettings() {
	// OOM score
	if c.OomScoreAdj != nil {
		if *c.OomScoreAdj <= c.client.container.HostConfig.OomScoreAdj {
			logWarning(
				"The controller's OOM score is", c.client.container.HostConfig.OomScoreAdj,
				"but the", c.Name, "component's is", *c.OomScoreAdj,
			)
			logWarning(
				"  This can potentially get the controller killed before the component.",
			)
		}
	}
}

func logWarning(v ...interface{}) {
	fmt.Print("[Warning] ")
	fmt.Println(v...)
}
