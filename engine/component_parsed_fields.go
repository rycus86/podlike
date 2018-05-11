package engine

import (
	"errors"
	"fmt"
	"github.com/docker/go-units"
)

func (c *Component) getOomScoreAdjust() int {
	if c.OomScoreAdj != nil {
		return *c.OomScoreAdj
	}

	score := c.client.container.HostConfig.OomScoreAdj

	if score < 1000 {
		// make it a little more likely to be killed than the controller
		return score + 1
	} else {
		return 1000
	}
}

func (c *Component) getMemoryLimit() (int64, error) {
	var (
		limit int64
		err   error
	)

	if c.MemoryLimit != "" {
		limit, err = units.RAMInBytes(c.MemoryLimit)

		if err != nil {
			return 0, err
		}
	} else {
		limit = 0
	}

	if c.client.container.HostConfig.Memory > 0 {
		if limit <= 0 || limit > c.client.container.HostConfig.Memory {
			return c.client.container.HostConfig.Memory, nil
		}
	}

	return limit, nil
}

func (c *Component) getMemorySwapLimit() (int64, error) {
	var (
		limit int64
		err   error
	)

	if c.MemorySwapLimit != "" {
		limit, err = units.RAMInBytes(c.MemorySwapLimit)

		if err != nil {
			return 0, err
		}
	} else {
		limit = 0
	}

	if c.client.container.HostConfig.MemorySwap > 0 {
		if limit <= 0 || limit > c.client.container.HostConfig.MemorySwap {
			return c.client.container.HostConfig.MemorySwap, nil
		}
	}

	return limit, nil
}

func (c *Component) getUlimits() ([]*units.Ulimit, error) {
	if c.Ulimits == nil {
		return nil, nil
	}

	res := make([]*units.Ulimit, len(c.Ulimits), len(c.Ulimits))
	index := 0

	for key, item := range c.Ulimits {
		var hard, soft int64

		if value, ok := item.(int64); ok {
			hard = value
			soft = value
		} else if value, ok := item.(int); ok {
			hard = int64(value)
			soft = int64(value)
		} else if mapped, ok := item.(map[interface{}]interface{}); ok {
			// TODO this could do with a bit more validation and error handling
			hardRaw, ok := mapped["hard"]
			if !ok {
				return nil, errors.New(fmt.Sprintf("'hard' not found in ulimit: %s = %+v %T", key, mapped, mapped))
			}

			softRaw, ok := mapped["soft"]
			if !ok {
				return nil, errors.New(fmt.Sprintf("'soft' not found in ulimit: %s = %+v %T", key, mapped, mapped))
			}

			if hard, ok = hardRaw.(int64); !ok {
				if hardInt, ok := hardRaw.(int); ok {
					hard = int64(hardInt)
				} else {
					return nil, errors.New(fmt.Sprintf("unexpected ulimit: %s.hard = %+v %T", key, hardRaw, hardRaw))
				}
			}

			if soft, ok = softRaw.(int64); !ok {
				if softInt, ok := softRaw.(int); ok {
					soft = int64(softInt)
				} else {
					return nil, errors.New(fmt.Sprintf("unexpected ulimit: %s.soft = %+v %T", key, softRaw, softRaw))
				}
			}
		} else {
			return nil, errors.New(fmt.Sprintf("unexpected ulimit: %s = %+v %T", key, item, item))
		}

		ulimit := units.Ulimit{
			Name: key,
			Hard: hard,
			Soft: soft,
		}

		res[index] = &ulimit

		index++
	}

	return res, nil
}
