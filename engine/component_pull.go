package engine

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io/ioutil"
)

func (c *Component) pullImage() error {
	fmt.Println("Pulling image:", c.Image)

	// TODO is context.Background() appropriate here?
	if reader, err := c.client.api.ImagePull(context.Background(), c.Image, types.ImagePullOptions{}); err != nil {
		return err
	} else {
		defer reader.Close()

		ioutil.ReadAll(reader)

		return nil
	}
}
