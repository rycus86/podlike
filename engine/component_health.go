package engine

import (
	"errors"
	"fmt"
	"github.com/rycus86/podlike/healthcheck"
)

func parseHealthcheckTest(value interface{}) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	stringValue, ok := value.(string)
	if ok {
		return []string{"CMD-SHELL", stringValue}, nil
	}

	sliceValue, ok := value.([]string)
	if ok {
		return sliceValue, nil
	}

	slice, ok := value.([]interface{})
	if ok {
		values := make([]string, len(slice), len(slice))

		for idx, item := range slice {
			values[idx] = fmt.Sprintf("%s", item)
		}

		return values, nil
	} else {
		return nil, errors.New(fmt.Sprintf("invalid string or slice: %T %+v", value, value))
	}
}

func (c *Component) initHealthCheckingIfNecessary() error {
	hasHealthcheck, err := c.hasHealthcheck()
	if err != nil {
		return err
	}

	if hasHealthcheck {
		healthcheck.Initialize(c.container.ID, healthcheck.StateStarting)
	}

	return nil
}

func (c *Component) hasHealthcheck() (bool, error) {
	return c.container.Config.Healthcheck != nil && c.container.Config.Healthcheck.Test != nil, nil
}
