package client

import "github.com/IDzetI/golang-test-task/pkg/service"

func (c *client) SetService(service service.Service) {
	c.service = service
}
