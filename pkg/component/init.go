package component

import (
	"fmt"
	"github.com/docker/go-units"
	"github.com/rycus86/podlike/pkg/api"
)

func (c *Component) Initialize(name string, client api.Controller, engine api.Engine) {
	c.Name = name
	c.client = client
	c.engine = engine

	c.warnForSettings()
}

func (c *Component) warnForSettings() {
	// Memory reservation
	if c.MemoryReservation != nil {
		logWarning("Memory reservation is set to", *c.MemoryReservation, "for component:", c.Name)
		logWarning("  For Swarm scheduling, it's probably better to set memory reservation on the service only.")
	}

	// Memory limit
	if c.client.GetHostConfig().Memory > 0 {
		if memLimit, err := units.RAMInBytes(c.MemoryLimit); err == nil {
			if memLimit > c.client.GetHostConfig().Memory {
				logWarning(
					"Memory limit on", c.Name, "is set to", memLimit, "but because of the controller,",
					"it's going to be overridden to", c.client.GetHostConfig().Memory,
				)
			}
		}
	}

	// Memory swap limit
	if c.client.GetHostConfig().MemorySwap > 0 {
		if memSwapLimit, err := units.RAMInBytes(c.MemorySwapLimit); err == nil {
			if memSwapLimit > c.client.GetHostConfig().MemorySwap {
				logWarning(
					"Memory swap limit on", c.Name, "is set to", memSwapLimit, "but because of the controller,",
					"it's going to be overridden to", c.client.GetHostConfig().MemorySwap,
				)
			}
		}
	}

	// OOM score
	if c.OomScoreAdj != nil {
		if *c.OomScoreAdj <= c.client.GetHostConfig().OomScoreAdj {
			logWarning(
				"The controller's OOM score is", c.client.GetHostConfig().OomScoreAdj,
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
