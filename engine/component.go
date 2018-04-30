package engine

func (c *Component) init(name string, client *Client) {
	c.Name = name
	c.client = client
}
